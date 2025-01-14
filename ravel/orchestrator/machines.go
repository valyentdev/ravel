package orchestrator

import (
	"context"
	"io"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
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
		return err
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

func (o *Orchestrator) WaitMachine(
	ctx context.Context,
	machine cluster.Machine,
	state api.MachineStatus,
	timeout uint,
) error {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	err = agentClient.WaitForMachineStatus(ctx, machine.Id, state, timeout)
	if errdefs.IsNotFound(err) {
		if state == api.MachineStatusDestroyed {
			return nil
		}
	}

	return err
}

func (o *Orchestrator) MachineExec(ctx context.Context, machine cluster.Machine, execOpts *api.ExecOptions) (*api.ExecResult, error) {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return nil, err
	}

	return agentClient.MachineExec(ctx, machine.Id, execOpts.Cmd, execOpts.GetTimeout())
}

func (o *Orchestrator) GetMachineLogsRaw(ctx context.Context, machine cluster.Machine, follow bool) (io.ReadCloser, error) {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return nil, err
	}

	return agentClient.GetMachineLogsRaw(ctx, machine.Id, follow)
}

func (o *Orchestrator) EnableMachineGateway(ctx context.Context, machine cluster.Machine) error {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	return agentClient.EnableMachineGateway(ctx, machine.Id)
}

func (o *Orchestrator) DisableMachineGateway(ctx context.Context, machine cluster.Machine) error {
	agentClient, err := o.getAgentClient(machine.Node)
	if err != nil {
		return err
	}

	return agentClient.DisableMachineGateway(ctx, machine.Id)
}
