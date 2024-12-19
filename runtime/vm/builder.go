package vm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/jailer"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"
	"github.com/valyentdev/ravel/runtime/images"
)

const (
	dataDir         = "/var/lib/ravel/instances"
	initRamfsPath   = "/initramfs"
	linuxKernelPath = "/vmlinux.bin"
	chAPISocketPath = "/instance.sock"
	vsockPath       = "/instance.vsock"
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

type Builder struct {
	cpuMhz       int64
	chBinary     string
	jailerBinary string
	initBinary   string
	initBin      []byte
	linuxKernel  string
	images       *images.Service
	ctrd         *client.Client
	snapshotter  string
	jailerUser   User
}

type User struct {
	Uid int
	Gid int
}

func NewBuilder(
	chBinary, jailerBinary, initBinary, linuxKernel string,
	images *images.Service,
	ctrd *client.Client,
	snapshotter string,
	cpuMhz int64,
	user User,
) (*Builder, error) {

	initBin, err := os.ReadFile(initBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to read init binary: %w", err)
	}

	_, err = cgroup2.NewManager("/sys/fs/cgroup", "/ravel", &cgroup2.Resources{}) // create cgroup2 manager
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup2 manager: %w", err)
	}

	return &Builder{
		chBinary:     chBinary,
		initBinary:   initBinary,
		linuxKernel:  linuxKernel,
		jailerBinary: jailerBinary,
		initBin:      initBin,
		images:       images,
		ctrd:         ctrd,
		snapshotter:  snapshotter,
		cpuMhz:       cpuMhz,
		jailerUser:   user,
	}, nil
}

var _ instance.Builder = (*Builder)(nil)

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

// BuildInstanceVM implements instance.VMBuilder.
func (b *Builder) BuildInstanceVM(ctx context.Context, instance *instance.Instance) (instance.VM, error) {
	startTime := time.Now()
	image, err := b.images.GetImage(ctx, instance.ImageRef)
	if err != nil {
		return nil, err
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

	vmConfig := b.getContainerMachineCHVmConfig(instance, rootfs)
	vm, err := newVM(instance.Id, cmd, vmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM: %w", err)
	}

	return vm, nil
}

// CleanupInstanceVM implements instance.VMBuilder.
func (b *Builder) CleanupInstanceVM(ctx context.Context, instance *instance.Instance) error {
	if err := b.removeRootFS(instance.Id); err != nil {
		return err
	}

	cg, err := cgroup2.Load("/ravel/" + instance.Id)
	if err != nil {
		return fmt.Errorf("failed to load cgroup: %w", err)
	}

	if err := cg.Delete(); err != nil {
		return fmt.Errorf("failed to delete cgroup: %w", err)
	}

	if err := os.RemoveAll(getInstanceDir(instance.Id)); err != nil {
		return err
	}
	return nil
}

func (b *Builder) CleanupInstance(ctx context.Context, instance *instance.Instance) error {
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

// RecoverInstanceVM implements instance.VMBuilder.
func (b *Builder) RecoverInstanceVM(ctx context.Context, i *instance.Instance) (instance.VM, error) {
	vmm, err := cloudhypervisor.NewVMMClient(getAPISocketPath(i.Id))
	if err != nil {
		return nil, err
	}
	vm := &vm{
		id:       i.Id,
		vmm:      vmm,
		vsock:    getVsockPath(i.Id),
		waitChan: make(chan struct{}),
	}

	ok := vm.recover()
	if !ok {
		return nil, errors.New("failed to recover VM")
	}

	return vm, nil
}

func (r *Builder) getContainerMachineCHVmConfig(i *instance.Instance, rootfs string) cloudhypervisor.VmConfig {
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
