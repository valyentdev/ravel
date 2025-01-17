package drivers

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/disks"
)

type InstanceTask interface {
	Start(ctx context.Context) error
	Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	Run() instance.ExitResult
	WaitExit(ctx context.Context) bool
	Signal(ctx context.Context, signal string) error
	Stop(ctx context.Context, signal string) error
	Shutdown(ctx context.Context) error
}

type Driver interface {
	BuildInstanceTask(ctx context.Context, instance *instance.Instance, disks []disks.Disk) (InstanceTask, error)
	CleanupInstanceTask(ctx context.Context, instance *instance.Instance) error
	RecoverInstanceTask(ctx context.Context, i *instance.Instance) (InstanceTask, error)
	CleanupInstance(ctx context.Context, instance *instance.Instance) error
	Snapshotter() string
}
