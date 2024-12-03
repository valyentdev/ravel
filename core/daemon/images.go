package daemon

import (
	"context"
	"time"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/registry"
)

type Image = images.Image

type ImagePullOptions struct {
	Ref  string                    `json:"ref"`
	Auth registry.RegistriesConfig `json:"auth,omitempty"`
}

type Runtime interface {
	CreateInstance(ctx context.Context, opt instance.InstanceOptions) (*instance.Instance, error)
	GetInstance(id string) (*instance.Instance, error)
	ListInstances() []instance.Instance
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string, opt *api.StopConfig) error
	GetInstanceLogs(id string) ([]*api.LogEntry, error)
	SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error)
	WaitInstanceExit(ctx context.Context, id string) (*instance.ExitResult, error)
	WatchInstanceState(ctx context.Context, id string) (<-chan instance.State, error)

	DeleteImage(ctx context.Context, ref string) error
	DestroyInstance(ctx context.Context, id string) error
	InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	ListImages(ctx context.Context) ([]images.Image, error)
	PruneImages(ctx context.Context) error
	PullImage(ctx context.Context, opt ImagePullOptions) (*images.Image, error)
}
