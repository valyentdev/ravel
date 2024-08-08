package container

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valyentdev/ravel/internal/agent/tap"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/cloudhypervisor"
)

func (r *Runtime) createRunningInstance(instance core.Instance) (*runningVM, error) {
	instanceId := instance.Id

	rootFS, err := r.prepareFilesystems(context.Background(), instance)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare instance filesystems for instance %q: %w", instanceId, err)
	}
	defer func() {
		if err != nil {
			r.fsBuilder.CleanupFilesystems(instanceId)
		}
	}()

	return NewRunningVM(instance, *r.getContainerMachineCHVmConfig(instance, rootFS))
}

func (r *Runtime) prepareFilesystems(ctx context.Context, instance core.Instance) (rootFs string, err error) {
	imageRef := instance.Config.Workload.Image

	image, err := r.images.GetImage(ctx, imageRef)
	if err != nil {
		return "", fmt.Errorf("failed to get image %q: %w", imageRef, err)
	}

	rootfs, err := r.fsBuilder.PrepareFilesystems(instance.Id, image)
	if err != nil {
		return "", fmt.Errorf("failed to prepare filesystem for instance %s: %w", instance.Id, err)
	}

	return rootfs, nil
}

func (r *Runtime) getContainerMachineCHVmConfig(i core.Instance, rootfs string) *cloudhypervisor.VmConfig {
	instanceId := i.Id
	config := i.Config
	return &cloudhypervisor.VmConfig{
		Cpus: &cloudhypervisor.CpusConfig{
			BootVcpus: int(config.Guest.VCpus),
			MaxVcpus:  int(config.Guest.VCpus),
		},
		Memory: &cloudhypervisor.MemoryConfig{
			Size: config.Guest.MemoryMB * 1_000_000,
		},
		Console: &cloudhypervisor.ConsoleConfig{
			Mode: "Pty",
		},
		Payload: cloudhypervisor.PayloadConfig{
			Initramfs: cloudhypervisor.StringPtr(getInitrdPath(instanceId)),
			Kernel:    cloudhypervisor.StringPtr(r.config.LinuxKernel),
			Cmdline:   cloudhypervisor.StringPtr("ro console=hvc0 rdinit=ravel-init"),
		},
		Disks: &[]cloudhypervisor.DiskConfig{
			{
				Path: rootfs,
			},
		},
		Net: &[]cloudhypervisor.NetConfig{
			{
				Tap: cloudhypervisor.StringPtr(tap.TapName(instanceId)),
			},
		},
		Vsock: &cloudhypervisor.VsockConfig{
			Cid:    3,
			Socket: vsockPath(instanceId),
		},
	}
}

func (r *Runtime) cleanupAfterVMRun(m *runningVM) {
	instanceId := m.Id()
	err := r.fsBuilder.CleanupFilesystems(instanceId)
	if err != nil {
		slog.Error("failed to cleanup filesystems", "err", err)
	}

}
