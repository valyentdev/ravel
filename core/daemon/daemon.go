package daemon

import (
	"context"
	"time"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/validation"
)

type InstanceOptions struct {
	Id     string                  `json:"id"`
	Config instance.InstanceConfig `json:"config"`
}

func (i *InstanceOptions) Validate() error {
	errDetails := validation.ValidateInstanceConfig(&i.Config, "body.config")

	err := validation.ValidateInstanceId(i.Id)
	if err != nil {
		errDetails = append(errDetails, &errdefs.ErrorDetail{
			Location: "body.id",
			Message:  err.Error(),
			Value:    i.Id,
		})
	}
	if len(errDetails) > 0 {
		return errdefs.NewInvalidArgument("instance config validation failed", errDetails...)
	}

	return nil
}

type Daemon interface {
	CreateInstance(ctx context.Context, opt InstanceOptions) (*instance.Instance, error)
	GetInstance(ctx context.Context, id string) (*instance.Instance, error)
	ListInstances(ctx context.Context) ([]instance.Instance, error)
	DestroyInstance(ctx context.Context, id string) error
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string, opt *api.StopConfig) error
	InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error)
	SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error)

	DeleteImage(ctx context.Context, ref string) error
	ListImages(ctx context.Context) ([]images.Image, error)
	PullImage(ctx context.Context, opt ImagePullOptions) (*images.Image, error)
}
