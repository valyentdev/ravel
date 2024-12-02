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

func (o *Orchestrator) PrepareAllocation(ctx context.Context, region string, allocationId string, resources api.Resources) (nodeId string, err error) {
	workers, err := o.broker.GetAvailableWorkers(placement.PlacementRequest{
		Region:       region,
		AllocationId: allocationId,
		Resources:    resources,
	})
	if err != nil {
		if err == placement.ErrPlacementFailed {
			slog.Warn("Failed to place machine", "machine_id", allocationId)
			err = errdefs.NewResourcesExhausted("failed to place machine")
		}
		return
	}

	candidate := workers[0]
	nodeId = candidate.NodeId
	return
}

func (o *Orchestrator) PutMachine(ctx context.Context, nodeId string, machine *cluster.Machine, mv api.MachineVersion, start bool) error {
	member, err := o.clusterState.GetNode(ctx, nodeId)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	ac, err := o.getAgentClient(member.Id)
	if err != nil {
		return fmt.Errorf("failed to get agent client: %w", err)
	}

	_, err = ac.PutMachine(ctx, cluster.PutMachineOptions{
		AllocationId: machine.Id,
		Machine:      *machine,
		Version:      mv,
		Start:        start,
	})
	if err != nil {
		return err
	}

	return nil
}
