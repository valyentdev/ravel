package orchestrator

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (o *Orchestrator) DestroyMachine(ctx context.Context, machine cluster.Machine, force bool) error {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.DestroyMachine(ctx, machine.Id, force)
	if err != nil {
		return err
	}

	return nil
}

func (o *Orchestrator) StartMachineInstance(ctx context.Context, machine cluster.Machine) error {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.StartMachine(ctx, machine.Id)
	if err != nil {
		return errdefs.NewUnknown("Failed to start machine")
	}

	return nil
}

func (o *Orchestrator) StopMachineInstance(ctx context.Context, machine cluster.Machine, stopConfig *api.StopConfig) error {
	slog.Debug("StopMachineInstance", "machine", machine)
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.StopMachine(ctx, machine.Id, stopConfig)
	if err != nil {
		slog.Debug("Failed to stop machine", "err", err)
		rerr, ok := err.(*errdefs.RavelError)
		if ok {
			for _, e := range rerr.Errors {
				slog.Debug("Error", "err", e)
			}
		}

		return errdefs.NewUnknown("Failed to stop machine")
	}

	return nil
}

type waitOpt struct {
	instanceId string
	timeout    time.Duration
}

type WaitOpt func(*waitOpt)

func WithInstanceId(instanceId string) WaitOpt {
	return func(o *waitOpt) {
		o.instanceId = instanceId
	}
}

func WithTimeout(timeout time.Duration) WaitOpt {
	return func(o *waitOpt) {
		o.timeout = timeout
	}
}

func (o *Orchestrator) WaitMachine(
	ctx context.Context,
	machine cluster.Machine,
	state api.MachineStatus,
	opts ...WaitOpt,
) error {
	opt := &waitOpt{
		instanceId: machine.InstanceId,
		timeout:    time.Second * 30,
	}

	timeoutCtx, cancelTimeoutCtx := context.WithTimeout(ctx, opt.timeout)
	defer cancelTimeoutCtx()

	for _, o := range opts {
		o(opt)
	}
	cancel, updates, err := o.clusterState.WatchInstance(ctx, machine.Id, opt.instanceId)
	if err != nil {
		return err
	}
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return errdefs.NewDeadlineExceeded("timeout reached while waiting for machine status")
		case update := <-updates:
			if update.Status == state {
				return nil
			}
		}
	}

}

func (o *Orchestrator) GetMachineLogsRaw(ctx context.Context, machine cluster.Machine, follow bool) (io.ReadCloser, error) {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return nil, err
	}

	return agentClient.GetMachineLogsRaw(ctx, machine.Id, follow)
}
