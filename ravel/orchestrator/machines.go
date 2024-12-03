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

type WaitOpt struct {
	InstanceId string
	Timeout    time.Duration
}

func (o *Orchestrator) WaitMachine(
	ctx context.Context,
	machineId string,
	state api.MachineStatus,
	opt WaitOpt,
) error {
	timeoutCtx, cancelTimeoutCtx := context.WithTimeout(ctx, opt.Timeout)
	defer cancelTimeoutCtx()

	slog.Info("Watching machine status", "machineId", machineId, "status", state)
	updates, err := o.clusterState.WatchInstanceStatus(timeoutCtx, machineId, opt.InstanceId)
	if err != nil {
		return err
	}
	t1 := time.Now()
	for {
		slog.Debug("waiting")
		select {
		case <-timeoutCtx.Done():
			slog.Debug("Timeout reached", "elapsed", time.Since(t1))
			return errdefs.NewDeadlineExceeded("timeout reached while waiting for machine status")
		case update := <-updates:
			slog.Debug("Machine status update", "machineId", machineId, "status", update)
			if api.MachineStatus(update) == state {
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
