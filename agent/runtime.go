package agent

import (
	"context"
	"time"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime"
)

func (a *Agent) DeleteImage(ctx context.Context, ref string) error {
	return a.runtime.DeleteImage(ctx, ref)
}

func (a *Agent) ListImages(ctx context.Context) ([]images.Image, error) {
	return a.runtime.ListImages(ctx)
}

func (a *Agent) PullImage(ctx context.Context, opts runtime.PullImageOptions) (*images.Image, error) {
	return a.runtime.PullImage(ctx, opts)
}
func (a *Agent) CreateInstance(ctx context.Context, opt structs.InstanceOptions) (*instance.Instance, error) {
	return a.runtime.CreateInstance(ctx, opt.Id, runtime.InstanceOptions{
		Config: opt.Config,
	})
}

func (s *Agent) StartInstance(ctx context.Context, id string) error {
	return s.runtime.StartInstance(ctx, id)
}

func (s *Agent) DestroyInstance(ctx context.Context, id string, force bool) error {
	return s.runtime.DestroyInstance(ctx, id)
}

func (s *Agent) StopInstance(ctx context.Context, id string, opt *api.StopConfig) error {
	return s.runtime.StopInstance(ctx, id, opt)
}

func (s *Agent) InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	return s.runtime.InstanceExec(ctx, id, cmd, timeout)
}

func (d *Agent) ListInstances(ctx context.Context) ([]instance.Instance, error) {
	return d.runtime.ListInstances(), nil
}

func (a *Agent) GetInstance(ctx context.Context, id string) (*instance.Instance, error) {
	return a.runtime.GetInstance(id)
}

func (a *Agent) SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	return a.runtime.SubscribeToInstanceLogs(ctx, id)
}

func (a *Agent) GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	return a.runtime.GetInstanceLogs(id)
}
