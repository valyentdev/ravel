package vm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/snapshots"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/jailer"
	initdclient "github.com/valyentdev/ravel/initd/client"
	"github.com/valyentdev/ravel/internal/resources"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"
	"github.com/valyentdev/ravel/runtime/drivers"
	"github.com/valyentdev/ravel/runtime/drivers/vm/tap"
)

const (
	dataDir         = "/var/lib/ravel/instances"
	initRamfsPath   = "/initramfs"
	linuxKernelPath = "/vmlinux.bin"
	chAPISocketPath = "/instance.sock"
	vsockPath       = "/instance.vsock"
	snapshotter     = "devmapper"
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

type Driver struct {
	cpuMhz       int64
	chBinary     string
	jailerBinary string
	initBinary   string
	linuxKernel  string
	ctrd         *client.Client
	jailerUser   User
	snapshotter  snapshots.Snapshotter
}

func (b *Driver) Snapshotter() string {
	return snapshotter
}

type User struct {
	Uid int
	Gid int
}

type Config struct {
	CloudHypervisorBinary string `json:"cloud_hypervisor_binary" toml:"cloud_hypervisor_binary"`
	JailerBinary          string `json:"jailer_binary" toml:"jailer_binary"`
	InitBinary            string `json:"init_binary" toml:"init_binary"`
	LinuxKernel           string `json:"linux_kernel" toml:"linux_kernel"`
}

func NewDriver(
	config Config,
	ctrd *client.Client,
) (*Driver, error) {
	frequency, err := resources.GetHostCPUFrequency()
	if err != nil {
		return nil, fmt.Errorf("failed to get host CPU frequency: %w", err)
	}

	uid, gid, err := setupRavelJailerUser()
	if err != nil {
		return nil, fmt.Errorf("failed to setup ravel jailer user: %w", err)
	}

	_, err = cgroup2.NewManager("/sys/fs/cgroup", "/ravel", &cgroup2.Resources{}) // create cgroup2 manager
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup2 manager: %w", err)
	}

	snapshotService := ctrd.SnapshotService(snapshotter)

	return &Driver{
		chBinary:     config.CloudHypervisorBinary,
		initBinary:   config.InitBinary,
		linuxKernel:  config.LinuxKernel,
		jailerBinary: config.JailerBinary,
		ctrd:         ctrd,
		snapshotter:  snapshotService,
		cpuMhz:       frequency,
		jailerUser: User{
			Uid: uid,
			Gid: gid,
		},
	}, nil
}

var _ drivers.Driver = (*Driver)(nil)

func getCgroupMemory(c *instance.InstanceGuestConfig) *cgroup2.Memory {
	high := int64(c.MemoryMB * 1_000_000)
	max := int64(float64(c.MemoryMB) * 1_000_000 * 1.1) // trigger OOM at 110% of the limit, should be tuned later
	return &cgroup2.Memory{
		High: &high,
		Max:  &max,
	}
}

func getCgroupCPU(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.CPU {
	quota := int64(float64(c.CpusMHz) / float64(cpuMhz) * 1_000_000)
	period := uint64(1_000_000)
	max := cgroup2.NewCPUMax(&quota, &period)
	return &cgroup2.CPU{
		Max: max,
	}
}

func getInstanceResources(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.Resources {
	return &cgroup2.Resources{
		CPU:    getCgroupCPU(cpuMhz, c),
		Memory: getCgroupMemory(c),
	}
}

// BuildInstanceTask implements instance.VMBuilder.
func (b *Driver) BuildInstanceTask(ctx context.Context, instance *instance.Instance) (drivers.InstanceTask, error) {
	_, err := tap.PrepareInstanceTapDevice(instance.Id, instance.Network, b.jailerUser.Uid, b.jailerUser.Gid)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare tap device: %w", err)
	}
	defer func() {
		if err != nil {
			err := tap.CleanupInstanceTapDevice(instance.Id, instance.Network)
			if err != nil {
				slog.Error("failed to cleanup tap device", "error", err)
			}
		}
	}()

	startTime := time.Now()
	image, err := b.ctrd.GetImage(ctx, instance.ImageRef)
	if err != nil {
		return nil, err
	}

	isUnpacked, err := image.IsUnpacked(ctx, b.Snapshotter())
	if err != nil {
		return nil, err
	}

	if !isUnpacked {
		err = image.Unpack(ctx, b.Snapshotter())
		if err != nil {
			return nil, err
		}
	}

	rootfs, err := b.prepareRootFS(ctx, instance.Id, image)
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

	_, err = cgroup.NewChild(instance.Id, getInstanceResources(b.cpuMhz, &instance.Config.Guest))
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup: %w", err)
	}

	jail, err := jailer.CreateJail(
		b.jailerBinary,
		jailer.JailConfig{
			Uid:     b.jailerUser.Uid,
			Gid:     b.jailerUser.Gid,
			NewRoot: getInstanceDir(instance.Id),
		},
		jailer.WithTUN(),
		jailer.WithKVM(),
		jailer.WithURandom(),
		jailer.WithBlockDevice(rootfs),
		jailer.WithBinary(b.chBinary, "/cloud-hypervisor"),
		jailer.WithHardLink(b.linuxKernel, "/vmlinux.bin", true),
		jailer.WithNewPidNS(),
		jailer.WithMountProc(),
		jailer.WithCgroup("/ravel/"+instance.Id),
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

	err = b.writeInitrd(initrd, instance, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to write initrd: %w", err)
	}

	slog.Debug("initrd created after", "time", time.Since(startTime))

	cmd := jail.Command("./cloud-hypervisor", "--api-socket", chAPISocketPath, "--log-file", "./cloud-hypervisor.log")

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	vmConfig := b.getContainerMachineCHVmConfig(instance, rootfs)
	vm := newVM(instance.Id, cmd, vmConfig)

	return vm, nil
}

// CleanupInstanceTask implements instance.VMBuilder.
func (b *Driver) CleanupInstanceTask(ctx context.Context, instance *instance.Instance) error {
	var errs []error

	if err := b.removeRootFS(instance.Id); err != nil {
		errs = append(errs, err)
	}

	if err := os.RemoveAll(getInstanceDir(instance.Id)); err != nil {
		errs = append(errs, err)
	}

	if err := tap.CleanupInstanceTapDevice(instance.Id, instance.Network); err != nil {
		errs = append(errs, err)
	}

	cg, err := cgroup2.Load("/ravel/" + instance.Id)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to load cgroup: %w", err))
	}

	if err := cg.Delete(); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete cgroup: %w", err))
	}

	return errors.Join(errs...)
}

func (b *Driver) CleanupInstance(ctx context.Context, instance *instance.Instance) error {
	err := b.removeRootFS(instance.Id)
	if err != nil {
		return fmt.Errorf("failed to remove rootfs: %w", err)
	}

	err = os.RemoveAll(getInstanceDir(instance.Id))
	if err != nil {
		return fmt.Errorf("failed to remove run dir: %w", err)
	}

	return nil
}

// RecoverInstanceTask implements instance.VMBuilder.
func (b *Driver) RecoverInstanceTask(ctx context.Context, i *instance.Instance) (drivers.InstanceTask, error) {
	vmm := cloudhypervisor.NewVMMClient(getAPISocketPath(i.Id))

	client := initdclient.NewInternalClient(getVsockPath(i.Id))

	vm := &vm{
		id:         i.Id,
		vmm:        vmm,
		waitChan:   make(chan struct{}),
		initClient: client,
	}

	ok := vm.recover()
	if !ok {
		return nil, errors.New("failed to recover VM")
	}

	return vm, nil
}

func (r *Driver) getContainerMachineCHVmConfig(i *instance.Instance, rootfs string) cloudhypervisor.VmConfig {
	config := i.Config
	return cloudhypervisor.VmConfig{
		Cpus: &cloudhypervisor.CpusConfig{
			BootVcpus: int(config.Guest.VCpus),
			MaxVcpus:  int(config.Guest.VCpus),
		},
		Memory: &cloudhypervisor.MemoryConfig{
			Size: int64(config.Guest.MemoryMB) * 1_000_000,
		},
		Console: &cloudhypervisor.ConsoleConfig{
			Mode: cloudhypervisor.ConsoleConfigModeTty,
		},
		Payload: cloudhypervisor.PayloadConfig{
			Initramfs: cloudhypervisor.StringPtr(initRamfsPath),
			Kernel:    cloudhypervisor.StringPtr(linuxKernelPath),
			Cmdline:   cloudhypervisor.StringPtr("ro console=hvc0 rdinit=ravel-init quiet"),
		},
		Disks: &[]cloudhypervisor.DiskConfig{
			{
				Path: rootfs,
			},
		},
		Net: &[]cloudhypervisor.NetConfig{
			{
				Tap: cloudhypervisor.StringPtr(i.Network.TapDevice),
			},
		},
		Vsock: &cloudhypervisor.VsockConfig{
			Cid:    3,
			Socket: vsockPath,
		},
	}
}
