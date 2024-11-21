package cluster

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
)

type PutMachineOptions struct {
	Machine      Machine
	Version      api.MachineVersion
	AllocationId string
	Start        bool
}

type Agent interface {
	PutMachine(ctx context.Context, opt PutMachineOptions) (*MachineInstance, error)
	StartMachine(ctx context.Context, machineId string) error
	StopMachine(ctx context.Context, machineId string, opt *api.StopConfig) error
	MachineExec(ctx context.Context, machineId string, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	DestroyMachine(ctx context.Context, machineId string, force bool) error
	SubscribeToMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error)
	GetMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, error)
}
