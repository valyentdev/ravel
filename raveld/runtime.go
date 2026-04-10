package raveld

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/core/daemon"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/core/validation"
	"github.com/alexisbouchez/ravel/runtime/disks"
	"github.com/containerd/containerd/v2/core/images"
)

func (a *Daemon) DeleteImage(ctx context.Context, ref string) error {
	return a.runtime.DeleteImage(ctx, ref)
}

func (a *Daemon) ListImages(ctx context.Context) ([]images.Image, error) {
	return a.runtime.ListImages(ctx)
}

func (a *Daemon) PullImage(ctx context.Context, opts daemon.ImagePullOptions) (*images.Image, error) {
	return a.runtime.PullImage(ctx, opts)
}

func (a *Daemon) CreateInstance(ctx context.Context, opt daemon.InstanceOptions) (*instance.Instance, error) {
	err := opt.Validate()
	if err != nil {
		return nil, err
	}

	network, err := a.network.AllocateNext()
	if err != nil {
		return nil, err
	}

	return a.runtime.CreateInstance(ctx, instance.InstanceOptions{
		Id:      opt.Id,
		Config:  opt.Config,
		Network: network,
	})
}

func (s *Daemon) StartInstance(ctx context.Context, id string) error {
	return s.runtime.StartInstance(ctx, id)
}

// StartInstanceFromSnapshot starts an instance by restoring from a snapshot.
// This enables sub-100ms cold starts for AI sandbox workloads.
func (s *Daemon) StartInstanceFromSnapshot(ctx context.Context, id string, globalSnapshotPath, jailSnapshotPath string) error {
	return s.runtime.StartInstanceFromSnapshot(ctx, id, globalSnapshotPath, jailSnapshotPath)
}

func (s *Daemon) DestroyInstance(ctx context.Context, id string) error {
	instance, err := s.runtime.GetInstance(id)
	if err != nil {
		return err
	}

	err = s.runtime.DestroyInstance(ctx, id)
	if err != nil {
		return err
	}

	if instance.Metadata.MachineId != "" {
		s.network.Release(instance.Network)
	}

	return err
}

func (s *Daemon) StopInstance(ctx context.Context, id string, opt *api.StopConfig) error {
	errs := validation.ValidateStopConfig(opt)
	if errs != nil {
		return errdefs.NewInvalidArgument("stop config validation failed", errs...)
	}
	return s.runtime.StopInstance(ctx, id, opt)
}

func (s *Daemon) InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	return s.runtime.InstanceExec(ctx, id, cmd, timeout)
}

func (d *Daemon) ListInstances(ctx context.Context) ([]instance.Instance, error) {
	return d.runtime.ListInstances(), nil
}

func (a *Daemon) GetInstance(ctx context.Context, id string) (*instance.Instance, error) {
	return a.runtime.GetInstance(id)
}

func (a *Daemon) SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	return a.runtime.SubscribeToInstanceLogs(ctx, id)
}

func (a *Daemon) GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	return a.runtime.GetInstanceLogs(id)
}

func (a *Daemon) CreateDisk(ctx context.Context, opts daemon.DiskOptions) (*disks.Disk, error) {
	return a.runtime.CreateDisk(ctx, opts.Id, opts.SizeMB)
}

func (a *Daemon) GetDisk(ctx context.Context, id string) (*disks.Disk, error) {
	return a.runtime.GetDisk(id)
}

func (a *Daemon) ListDisks(ctx context.Context) ([]disks.Disk, error) {
	return a.runtime.ListDisks()
}

func (a *Daemon) DestroyDisk(ctx context.Context, id string) error {
	return a.runtime.DestroyDisk(id)
}

// InstanceSnapshot saves the VM state for fast restore.
func (a *Daemon) InstanceSnapshot(ctx context.Context, id string, path string) error {
	return a.runtime.InstanceSnapshot(ctx, id, path)
}

// InstanceRestore restores the VM from a previously saved snapshot.
func (a *Daemon) InstanceRestore(ctx context.Context, id string, path string) error {
	return a.runtime.InstanceRestore(ctx, id, path)
}
