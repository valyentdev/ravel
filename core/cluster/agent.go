package cluster

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
)

type PutMachineOptions struct {
	Machine       Machine            `json:"machine"`
	Version       api.MachineVersion `json:"version"`
	AllocationId  string             `json:"allocation_id"`
	Start         bool               `json:"start"`
	EnableGateway bool               `json:"enable_gateway"`
}

type Agent interface {
	// PutMachine confirm an allocation placed before on the agent
	// It returns the machine instance created on the agent
	// The agent should then create an instance of the machine and start it
	// if the start flag is set to true
	PutMachine(ctx context.Context, opt PutMachineOptions) (*MachineInstance, error)
	StartMachine(ctx context.Context, machineId string) error
	StopMachine(ctx context.Context, machineId string, opt *api.StopConfig) error
	MachineExec(ctx context.Context, machineId string, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	DestroyMachine(ctx context.Context, machineId string, force bool) error
	SubscribeToMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error)
	GetMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, error)

	EnableMachineGateway(ctx context.Context, id string) error
	DisableMachineGateway(ctx context.Context, id string) error
}
