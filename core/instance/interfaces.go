package instance

import (
	"context"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/api"
)

type EventReporter interface {
	ReportInstanceEvent(event Event)
}

type InstanceStore interface {
	LoadInstances() ([]Instance, error)
	PutInstance(instance Instance) error
	DeleteInstance(id string) error
}

type Handle struct {
	Console string
}

type ExitResult = InstanceExitedEventPayload

type VM interface {
	Start(ctx context.Context) (Handle, error)
	Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	Run() *ExitResult
	WaitExit(ctx context.Context) bool
	Signal(ctx context.Context, signal string) error
	Stop(ctx context.Context, signal string) error
	Shutdown(ctx context.Context) error
}

type Builder interface {
	PrepareInstance(ctx context.Context, instance *Instance, image client.Image) error
	BuildInstanceVM(ctx context.Context, instance *Instance) (VM, error)
	RecoverInstanceVM(ctx context.Context, instance *Instance) (VM, Handle, error)
	CleanupInstanceVM(ctx context.Context, instance *Instance) error
	CleanupInstance(ctx context.Context, instance *Instance) error
}

type NetworkingService interface {
	EnsureInstanceNetwork(id string, config NetworkingConfig) error
	CleanupInstanceNetwork(id string, config NetworkingConfig) error
}
