package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/agent/machinerunner"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
)

func (a *Agent) onMachineDestroyed(m structs.MachineInstance) {
	a.machines.RemoveMachine(m.Machine.Id)
	err := a.store.DeleteMachineInstance(m.Machine.Id)
	if err != nil {
		slog.Error("failed to delete machine instance", "machine_id", m.Machine.Id, "err", err)
	}

	err = a.allocator.DeleteAllocation(m.Machine.Id)
	if err != nil {
		slog.Error("failed to release reservation", "machine_id", m.Machine.Id, "err", err)
	}
}

func (a *Agent) reportState(mi cluster.MachineInstance) error {
	return a.cluster.UpsertInstance(context.Background(), mi)
}

func (a *Agent) newMachine(machineInstance structs.MachineInstance) *machinerunner.MachineRunner {
	return machinerunner.New(
		a.store,
		machineInstance,
		a.runtime,
		a.reportState,
		a.eventer,
		a.onMachineDestroyed,
	)
}

func (a *Agent) PutMachine(ctx context.Context, opt cluster.PutMachineOptions) (*cluster.MachineInstance, error) {
	_, err := a.allocator.ConfirmAllocation(opt.AllocationId)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm reservation: %w", err)
	}

	defer func() {
		if err != nil {
			if err := a.allocator.DeleteAllocation(opt.AllocationId); err != nil {
				slog.Error("failed to release reservation", "err", err)
			}
		}
	}()

	var desiredStatus api.MachineStatus
	if opt.Start {
		desiredStatus = api.MachineStatusRunning
	} else {
		desiredStatus = api.MachineStatusStopped
	}

	machineInstance := structs.MachineInstance{
		Machine: opt.Machine,
		Version: opt.Version,
		State: structs.MachineInstanceState{
			DesiredStatus: desiredStatus,
			Status:        api.MachineStatusCreated,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	if err := a.store.CreateMachineInstance(machineInstance); err != nil {
		return nil, fmt.Errorf("failed to put machine: %w", err)
	}

	machine := a.newMachine(machineInstance)
	a.machines.AddMachine(machine)
	go machine.Run()

	ci := machineInstance.ClusterInstance()

	return &ci, nil
}

func (d *Agent) DestroyMachine(ctx context.Context, id string, force bool) error {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		return err
	}

	return machine.Destroy(ctx, force)
}

func (d *Agent) StartMachine(ctx context.Context, id string) error {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		return err
	}

	return machine.Start(ctx)
}

func (d *Agent) StopMachine(ctx context.Context, id string, opt *api.StopConfig) error {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		slog.Error("failed to get machine", "machine_id", id, "error", err)
		return err
	}

	return machine.Stop(ctx, opt)
}

func (d *Agent) SubscribeToMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		return nil, nil, err
	}

	return machine.SubscribeToLogs(ctx, id)
}

func (d *Agent) GetMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		return nil, err
	}

	return machine.GetLogs()
}

func (d *Agent) MachineExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	machine, err := d.machines.GetMachine(id)
	if err != nil {
		return nil, err
	}

	return machine.Exec(ctx, cmd, timeout)
}
