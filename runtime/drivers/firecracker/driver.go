package firecracker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/core/jailer"
	"github.com/alexisbouchez/ravel/internal/resources"
	"github.com/alexisbouchez/ravel/runtime/disks"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/alexisbouchez/ravel/runtime/drivers/common"
	"github.com/alexisbouchez/ravel/runtime/drivers/vm/tap"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/containerd/errdefs"
	"github.com/opencontainers/image-spec/identity"
)

const (
	dataDir         = "/var/lib/ravel/instances"
	initRamfsPath   = "/initramfs"
	linuxKernelPath = "/vmlinux.bin"
	fcAPISocketPath = "/instance.sock"
	vsockPath       = "/instance.vsock"
	snapshotterName = "devmapper"
)

func getInstanceDir(id string) string {
	return path.Join(dataDir, id)
}

func getAPISocketPath(id string) string {
	return path.Join(getInstanceDir(id), "instance.sock")
}

func getVsockPath(id string) string {
	return path.Join(getInstanceDir(id), "instance.vsock")
}

// Driver implements the drivers.Driver interface for Firecracker.
type Driver struct {
	cpuMhz            int64
	firecrackerBinary string
	jailerBinary      string
	initBinary        string
	linuxKernel       string
	ctrd              *client.Client
	jailerUser        common.User
	snapshotter       snapshots.Snapshotter
}

func (d *Driver) Snapshotter() string {
	return snapshotterName
}

// Config holds configuration for the Firecracker driver.
type Config struct {
	FirecrackerBinary string `json:"firecracker_binary" toml:"firecracker_binary"`
	JailerBinary      string `json:"jailer_binary" toml:"jailer_binary"`
	InitBinary        string `json:"init_binary" toml:"init_binary"`
	LinuxKernel       string `json:"linux_kernel" toml:"linux_kernel"`
}

// NewDriver creates a new Firecracker driver.
func NewDriver(config Config, ctrd *client.Client) (*Driver, error) {
	frequency, err := resources.GetHostCPUFrequency()
	if err != nil {
		return nil, fmt.Errorf("failed to get host CPU frequency: %w", err)
	}

	uid, gid, err := common.SetupRavelJailerUser()
	if err != nil {
		return nil, fmt.Errorf("failed to setup ravel jailer user: %w", err)
	}

	_, err = cgroup2.NewManager("/sys/fs/cgroup", "/ravel", &cgroup2.Resources{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup2 manager: %w", err)
	}

	snapshotService := ctrd.SnapshotService(snapshotterName)

	return &Driver{
		firecrackerBinary: config.FirecrackerBinary,
		initBinary:        config.InitBinary,
		linuxKernel:       config.LinuxKernel,
		jailerBinary:      config.JailerBinary,
		ctrd:              ctrd,
		snapshotter:       snapshotService,
		cpuMhz:            frequency,
		jailerUser: common.User{
			Uid: uid,
			Gid: gid,
		},
	}, nil
}

var _ drivers.Driver = (*Driver)(nil)

// BuildInstanceTask implements drivers.Driver.
func (d *Driver) BuildInstanceTask(ctx context.Context, inst *instance.Instance, instanceDisks []disks.Disk) (drivers.InstanceTask, error) {
	_, err := tap.PrepareInstanceTapDevice(inst.Id, inst.Network, d.jailerUser.Uid, d.jailerUser.Gid)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tap device: %w", err)
	}
	defer func() {
		if err != nil {
			if cleanupErr := tap.CleanupInstanceTapDevice(inst.Id, inst.Network); cleanupErr != nil {
				slog.Error("failed to cleanup tap device", "error", cleanupErr)
			}
		}
	}()

	startTime := time.Now()
	image, err := d.ctrd.GetImage(ctx, inst.ImageRef)
	if err != nil {
		return nil, err
	}

	isUnpacked, err := image.IsUnpacked(ctx, d.Snapshotter())
	if err != nil {
		return nil, err
	}

	if !isUnpacked {
		err = image.Unpack(ctx, d.Snapshotter())
		if err != nil {
			return nil, err
		}
	}

	rootfs, err := d.prepareRootFS(ctx, inst.Id, image)
	if err != nil {
		return nil, err
	}

	slog.Debug("rootfs prepared after", "time", time.Since(startTime))

	spec, err := image.Spec(ctx)
	if err != nil {
		return nil, err
	}

	cgroup, err := cgroup2.Load("/ravel")
	if err != nil {
		return nil, fmt.Errorf("failed to load cgroup: %w", err)
	}

	_, err = cgroup.NewChild(inst.Id, common.GetInstanceResources(d.cpuMhz, &inst.Config.Guest))
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup: %w", err)
	}

	opts := []jailer.Opt{
		jailer.WithTUN(),
		jailer.WithKVM(),
		jailer.WithURandom(),
		jailer.WithBlockDevice(rootfs),
		jailer.WithBinary(d.firecrackerBinary, "/firecracker"),
		jailer.WithHardLink(d.linuxKernel, "/vmlinux.bin", true),
		jailer.WithNewPidNS(),
		jailer.WithMountProc(),
		jailer.WithCgroup("/ravel/" + inst.Id),
	}
	for _, disk := range instanceDisks {
		opts = append(opts, jailer.WithBlockDevice(disk.Path))
	}

	jail, err := jailer.CreateJail(
		d.jailerBinary,
		jailer.JailConfig{
			Uid:     d.jailerUser.Uid,
			Gid:     d.jailerUser.Gid,
			NewRoot: getInstanceDir(inst.Id),
		},
		opts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create jail: %w", err)
	}

	slog.Debug("jail created after", "time", time.Since(startTime))

	initrd, err := jail.CreateFile(initRamfsPath, 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to create initrd: %w", err)
	}
	defer initrd.Close()

	err = common.WriteInitrd(initrd, d.initBinary, inst, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to write initrd: %w", err)
	}

	slog.Debug("initrd created after", "time", time.Since(startTime))

	cmd := jail.Command("./firecracker", "--api-sock", fcAPISocketPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	vmConfig := d.getVMConfig(inst, rootfs, instanceDisks)
	vm := newVM(inst.Id, cmd, vmConfig)

	return vm, nil
}

// CleanupInstanceTask implements drivers.Driver.
func (d *Driver) CleanupInstanceTask(ctx context.Context, inst *instance.Instance) error {
	var errs []error

	if err := d.removeRootFS(inst.Id); err != nil {
		errs = append(errs, err)
	}

	if err := os.RemoveAll(getInstanceDir(inst.Id)); err != nil {
		errs = append(errs, err)
	}

	if err := tap.CleanupInstanceTapDevice(inst.Id, inst.Network); err != nil {
		errs = append(errs, err)
	}

	cg, err := cgroup2.Load("/ravel/" + inst.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to load cgroup: %w", err))
	} else {
		if err := cg.Delete(); err != nil {
			errs = append(errs, fmt.Errorf("failed to delete cgroup: %w", err))
		}
	}

	return errors.Join(errs...)
}

// CleanupInstance implements drivers.Driver.
func (d *Driver) CleanupInstance(ctx context.Context, inst *instance.Instance) error {
	err := d.removeRootFS(inst.Id)
	if err != nil {
		return fmt.Errorf("failed to remove rootfs: %w", err)
	}

	err = os.RemoveAll(getInstanceDir(inst.Id))
	if err != nil {
		return fmt.Errorf("failed to remove run dir: %w", err)
	}

	return nil
}

// RecoverInstanceTask implements drivers.Driver.
func (d *Driver) RecoverInstanceTask(ctx context.Context, inst *instance.Instance) (drivers.InstanceTask, error) {
	vm := &firecrackerVM{
		id:       inst.Id,
		waitChan: make(chan struct{}),
	}

	ok := vm.recover()
	if !ok {
		return nil, errors.New("failed to recover VM")
	}

	return vm, nil
}

// RootFS helpers

func rootFSName(id string) string {
	return fmt.Sprintf("%s-%s", id, "rootfs")
}

func (d *Driver) removeRootFS(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := d.snapshotter.Remove(ctx, rootFSName(id))
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to remove snapshot %q: %w", id, err)
	}

	return nil
}

func (d *Driver) prepareRootFS(ctx context.Context, id string, image client.Image) (rootfs string, err error) {
	slog.Debug("preparing rootfs for container", "id", id)
	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get rootfs for image %q: %w", image.Name(), err)
	}

	parent := identity.ChainID(diffIDs).String()

	slog.Debug("preparing snapshot", "id", id, "parent", parent)

	labels := map[string]string{
		"containerd.io/gc.root": time.Now().UTC().Format(time.RFC3339),
	}

	mounts, err := d.snapshotter.Prepare(ctx, rootFSName(id), parent, snapshots.WithLabels(labels))
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			return "", fmt.Errorf("failed to prepare snapshot %q: %w", id, err)
		}

		slog.Debug("snapshot already exists, removing", "id", id)
		err = d.removeRootFS(id)
		if err != nil {
			return "", fmt.Errorf("failed to remove existing snapshot %q: %w", id, err)
		}

		slog.Debug("retrying snapshot preparation", "id", id, "parent", parent)
		mounts, err = d.snapshotter.Prepare(context.Background(), rootFSName(id), parent, snapshots.WithLabels(labels))
		if err != nil {
			return "", fmt.Errorf("failed to prepare snapshot %q: %w", id, err)
		}
	}

	if len(mounts) == 0 {
		return "", fmt.Errorf("no mounts found for container %q", id)
	}

	return mounts[0].Source, nil
}
