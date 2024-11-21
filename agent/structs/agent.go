package structs

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime"
)

type InstanceOptions struct {
	Id     string                  `json:"id"`
	Config instance.InstanceConfig `json:"config"`
}

type Runtime interface {
	CreateInstance(ctx context.Context, options InstanceOptions) (*instance.Instance, error)
	ListInstances(ctx context.Context) ([]instance.Instance, error)
	GetInstance(ctx context.Context, id string) (*instance.Instance, error)
	DestroyInstance(ctx context.Context, id string, force bool) error
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string, opt *api.StopConfig) error
	InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error)
	GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error)
	ListImages(ctx context.Context) ([]runtime.Image, error)
	PullImage(ctx context.Context, opts runtime.PullImageOptions) (*runtime.Image, error)
	DeleteImage(ctx context.Context, ref string) error
}

type Agent interface {
	Runtime
	cluster.Agent
}
