package runtimes

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
)

type NetworkConfig struct {
	LocalConfig networking.LocalConfig
}

type Handle struct {
	Console string
}
type Runtime interface {
	PrepareInstance(ctx context.Context, instance core.Instance, networkConfig NetworkConfig) (err error, fatal bool)
	RecoverVM(ctx context.Context, instance core.Instance) (Handle, bool)
	StartVM(ctx context.Context, instance core.Instance) (Handle, error)
	Exec(ctx context.Context, instanceId string, cmd []string, timeout time.Duration) (*ExecResult, error)
	SignalVM(ctx context.Context, instanceId string, signal string) error
	WaitVM(ctx context.Context, instanceId string) (*ExitResult, error)
	StopVM(ctx context.Context, instanceId string, signal string, timeout time.Duration) error
	DestroyVM(ctx context.Context, instanceId string) error
	DestroyInstance(ctx context.Context, instanceId string) error
}

type ExecResult struct {
	Stderr   []byte
	Stdout   []byte
	ExitCode int
}

type ExitResult = core.InstanceExitedEventPayload
