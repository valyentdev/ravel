package core

import (
	"context"
	"time"
)

type CreateInstancePayload struct {
	MachineId      string         `json:"machine_id"`
	MachineVersion string         `json:"machine_version"`
	FleetId        string         `json:"fleet_id"`
	Namespace      string         `json:"namespace"`
	Config         InstanceConfig `json:"config"`
	Start          bool           `json:"start"`
}

type InstanceExecOptions struct {
	Timeout *time.Duration `json:"timeout,omitempty"`
	Cmd     []string       `json:"cmd"`
}

func (e *InstanceExecOptions) GetTimeout() time.Duration {
	if e.Timeout == nil {
		return 5 * time.Second
	}
	return *e.Timeout
}

type ExecResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type Agent interface {
	CreateInstance(ctx context.Context, options CreateInstancePayload) (*Instance, error)
	ListInstances(ctx context.Context) ([]Instance, error)
	GetInstance(ctx context.Context, id string) (*Instance, error)
	DestroyInstance(ctx context.Context, id string, force bool) error
	StartInstance(ctx context.Context, id string) error
	StopInstance(ctx context.Context, id string, opt *StopConfig) error
	InstanceExec(ctx context.Context, id string, opt InstanceExecOptions) (*ExecResult, error)
}
