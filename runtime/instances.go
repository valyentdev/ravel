package runtime

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/runtime/instancerunner"
)

func (r *Runtime) StartInstance(ctx context.Context, id string) error {
	instance, err := r.getInstance(id)
	if err != nil {
		return err
	}

	err = instance.Start(context.Background())
	if err != nil {
		return err
	}

	return nil
}

// StartInstanceFromSnapshot starts the instance by restoring from a snapshot.
// This enables sub-100ms cold starts for AI sandbox workloads.
// globalSnapshotPath is the path to the global snapshot storage
// jailSnapshotPath is the jail-relative path where the snapshot will be placed
func (r *Runtime) StartInstanceFromSnapshot(ctx context.Context, id string, globalSnapshotPath, jailSnapshotPath string) error {
	instance, err := r.getInstance(id)
	if err != nil {
		return err
	}

	err = instance.StartFromSnapshot(context.Background(), globalSnapshotPath, jailSnapshotPath)
	if err != nil {
		return err
	}

	return nil
}

func (r *Runtime) DestroyInstance(ctx context.Context, id string) error {
	instance, err := r.instances.GetInstance(id)
	if err != nil {
		return err
	}

	err = instance.Destroy(context.Background())
	if err != nil {
		return err
	}

	err = r.disks.DetachInstance(instance.Instance().Config.GetDisks()...)
	if err != nil {
		return err
	}

	err = r.instancesStore.DeleteInstance(id)
	if err != nil {
		return err
	}

	r.instances.Delete(id)
	r.instances.ReleaseId(id)
	r.imagesUsage.ReleaseImage(instance.Instance().ImageRef)

	return nil
}

func (r *Runtime) StopInstance(ctx context.Context, id string, opt *api.StopConfig) error {
	instance, err := r.getInstance(id)
	if err != nil {
		return nil
	}

	err = instance.Stop(context.Background(), opt)
	if err != nil {
		return nil
	}

	return nil
}

func (r *Runtime) InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	i, err := r.getInstance(id)
	if err != nil {
		return nil, err
	}

	res, err := i.Exec(ctx, cmd, timeout)
	if err != nil {
		return nil, err
	}

	return &api.ExecResult{
		Stdout:   res.Stdout,
		ExitCode: res.ExitCode,
	}, nil
}

func (r *Runtime) ListInstances() []instance.Instance {
	return r.instances.List()
}

func (r *Runtime) GetInstance(id string) (*instance.Instance, error) {
	ir, err := r.getInstance(id)
	if err != nil {
		return nil, err
	}
	i := ir.Instance()

	return &i, nil
}

func (r *Runtime) getInstance(id string) (*instancerunner.InstanceRunner, error) {
	return r.instances.GetInstance(id)
}

func (r *Runtime) SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	ir, err := r.getInstance(id)
	if err != nil {
		return nil, nil, err
	}

	replay, sub := ir.SubscribeToLogs()

	ch := sub.Ch()

	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()

	return replay, ch, nil
}

func (r *Runtime) GetInstanceLogs(id string) ([]*api.LogEntry, error) {
	ir, err := r.getInstance(id)
	if err != nil {
		return nil, err
	}

	return ir.GetLog(), nil
}

func (r *Runtime) WatchInstanceState(ctx context.Context, id string) (<-chan instance.State, error) {
	ir, err := r.getInstance(id)
	if err != nil {
		return nil, err
	}

	return ir.WatchState(ctx), nil
}

// InstanceSnapshot saves the VM state for fast restore.
// This enables sub-100ms cold starts for AI sandbox workloads.
func (r *Runtime) InstanceSnapshot(ctx context.Context, id string, path string) error {
	ir, err := r.getInstance(id)
	if err != nil {
		return err
	}

	return ir.Snapshot(ctx, path)
}

// InstanceRestore restores the VM from a previously saved snapshot.
func (r *Runtime) InstanceRestore(ctx context.Context, id string, path string) error {
	ir, err := r.getInstance(id)
	if err != nil {
		return err
	}

	return ir.Restore(ctx, path)
}
