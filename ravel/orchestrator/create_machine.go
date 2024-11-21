package orchestrator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/cluster/placement"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (m *Orchestrator) PlaceMachine(ctx context.Context, machine *cluster.Machine, mv api.MachineVersion, start bool) error {
	machineId := machine.Id

	workers, err := m.broker.GetAvailableWorkers(placement.PlacementRequest{
		Region:       machine.Region,
		AllocationId: machineId,
		Resources:    mv.Resources,
	})
	if err != nil {
		if err == placement.ErrPlacementFailed {
			slog.Warn("Failed to place machine", "machine_id", machineId)
			err = errdefs.NewResourcesExhausted("failed to place machine")
		}
		return err
	}

	candidate := workers[0]

	member, err := m.clusterState.GetNode(ctx, candidate.NodeId)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	machine.Node = member.Id

	ac, err := m.getAgentClient(member.Id)
	if err != nil {
		return fmt.Errorf("failed to get agent client: %w", err)
	}

	_, err = ac.PutMachine(ctx, cluster.PutMachineOptions{
		AllocationId: machineId,
		Machine:      *machine,
		Version:      mv,
		Start:        start,
	})
	if err != nil {
		return err
	}

	return nil
}
