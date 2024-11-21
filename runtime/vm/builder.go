package vm

import (
	"context"
	"errors"
	"fmt"
	"path"
	"syscall"

	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/core/images"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/pkg/cloudhypervisor"
)

type Builder struct {
	runPath     Dir
	dataPath    Dir
	initBinary  string
	linuxKernel string
	images      *images.Service
	ctrd        *client.Client
	snapshotter string
}

func NewBuilder(
	runPath, dataPath Dir,
	initBinary, linuxKernel string,
	images *images.Service,
	ctrd *client.Client,
	snapshotter string,
) *Builder {
	return &Builder{
		runPath:     runPath,
		dataPath:    dataPath,
		initBinary:  initBinary,
		linuxKernel: linuxKernel,
		images:      images,
		ctrd:        ctrd,
		snapshotter: snapshotter,
	}
}

func (b *Builder) getInitRDPath(instanceId string) string {
	return path.Join(b.dataPath.InstanceDir(instanceId), "initrd")
}

var _ instance.Builder = (*Builder)(nil)

func (b *Builder) PrepareInstance(ctx context.Context, i *instance.Instance, image client.Image) error {
	if err := createInstanceDirectories(b.dataPath, b.runPath, i.Id); err != nil {
		return err
	}

	spec, err := image.Spec(ctx)
	if err != nil {
		return err
	}

	return b.prepareInitRD(i, spec)
}

// BuildInstanceVM implements instance.VMBuilder.
func (b *Builder) BuildInstanceVM(ctx context.Context, instance *instance.Instance) (instance.VM, error) {
	image, err := b.images.GetImage(ctx, instance.ImageRef)
	if err != nil {
		return nil, err
	}

	rootfs, err := b.prepareRootFS(ctx, instance.Id, image)
	if err != nil {
		return nil, err
	}
	vmConfig := b.getContainerMachineCHVmConfig(instance, rootfs)
	vm, err := newVM(instance.Id, vmConfig, b.socketPath(instance.Id), b.vsockPath(instance.Id))
	if err != nil {
		return nil, err
	}
	return vm, nil
}

// CleanupInstanceVM implements instance.VMBuilder.
func (b *Builder) CleanupInstanceVM(ctx context.Context, instance *instance.Instance) error {
	if err := b.removeRootFS(instance.Id); err != nil {
		return err
	}
	return nil
}

func (b *Builder) CleanupInstance(ctx context.Context, instance *instance.Instance) error {
	if err := removeInstanceDirectories(b.dataPath, b.runPath, instance.Id); err != nil {
		return err
	}
	return nil
}

// RecoverInstanceVM implements instance.VMBuilder.
func (b *Builder) RecoverInstanceVM(ctx context.Context, i *instance.Instance) (instance.VM, instance.Handle, error) {
	var h instance.Handle
	vmm, err := cloudhypervisor.NewVMM(
		b.socketPath(i.Id),
		cloudhypervisor.WithSysProcAttr(&syscall.SysProcAttr{
			Setsid: true,
		}),
	)
	if err != nil {
		return nil, h, err
	}

	vm := &vm{
		id:       i.Id,
		vmm:      vmm,
		vsock:    b.vsockPath(i.Id),
		waitChan: make(chan struct{}),
	}

	h, ok := vm.recover()
	if !ok {
		return nil, h, errors.New("failed to recover VM")
	}

	return vm, h, nil
}

func (r *Builder) getContainerMachineCHVmConfig(i *instance.Instance, rootfs string) cloudhypervisor.VmConfig {
	instanceId := i.Id
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
			Mode: "Pty",
		},
		Payload: cloudhypervisor.PayloadConfig{
			Initramfs: cloudhypervisor.StringPtr(r.getInitRDPath(instanceId)),
			Kernel:    cloudhypervisor.StringPtr(r.linuxKernel),
			Cmdline:   cloudhypervisor.StringPtr("ro console=hvc0 rdinit=ravel-init"),
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
			Socket: r.vsockPath(instanceId),
		},
	}
}

func (r *Builder) vsockPath(instanceId string) string {
	return fmt.Sprintf("/tmp/%s-vsock.sock", instanceId)
}

func (r *Builder) socketPath(instanceId string) string {
	return path.Join(r.runPath.InstanceDir(instanceId), "instance.sock")
}
