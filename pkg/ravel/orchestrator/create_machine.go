package orchestrator

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valyentdev/ravel/internal/placement"
	"github.com/valyentdev/ravel/pkg/core"
)

type CreateInstanceOptions struct {
	Machine   core.Machine        `json:"machine"`
	Config    core.InstanceConfig `json:"config"`
	SkipStart bool                `json:"skip_start"`
}

func (m *Orchestrator) CreateInstanceForMachine(ctx context.Context, namespace string, fleetId string, options CreateInstanceOptions) (*core.Instance, error) {
	machineId := options.Machine.Id

	workers, err := m.broker.GetAvailableWorkers(placement.MachinePlacementRequest{
		Region:        options.Machine.Region,
		ReservationId: machineId,
		Cpus:          int(options.Config.Guest.Cpus),
		Memory:        int(options.Config.Guest.MemoryMB),
	})
	if err != nil {
		if err == placement.ErrPlacementFailed {
			slog.Warn("Failed to place machine", "machine_id", machineId)
			err = core.NewResourcesExhausted("failed to place machine")
		}
		return nil, err
	}

	candidate := workers[0]

	member, err := m.clusterState.GetNode(context.Background(), candidate.NodeId)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	agentClient, err := m.getAgentClient(member.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent client: %w", err)
	}

	instance, err := agentClient.CreateInstance(context.Background(), core.CreateInstancePayload{
		MachineId: machineId,
		Namespace: namespace,
		FleetId:   fleetId,
		Start:     !options.SkipStart,
		Config:    options.Config,
	})
	if err != nil {
		return nil, err
	}

	return instance, nil
}
