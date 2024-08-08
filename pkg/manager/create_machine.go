package manager

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/internal/placement"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/proto"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CreateMachineOptions struct {
	Region    string             `json:"region"`
	Config    core.MachineConfig `json:"config"`
	SkipStart bool               `json:"skip_start"`
}

func (m *Manager) CreateMachine(ctx context.Context, namespace string, fleetId string, options CreateMachineOptions) error {
	machineId := id.Generate()
	createdAt := time.Now()

	workers, err := m.broker.GetAvailableWorkers(placement.MachinePlacementRequest{
		Region:        options.Region,
		ReservationId: machineId,
		Cpus:          int(options.Config.Guest.Cpus),
		Memory:        int(options.Config.Guest.MemoryMB),
	})
	if err != nil {
		if err == placement.ErrPlacementFailed {
			slog.Warn("Failed to place machine", "machine_id", machineId)
			return ravelerrors.NewResourcesExhausted("failed to place machine")
		}
		return err
	}

	candidate := workers[0]

	slog.Info("Creating machine", "machine_id", machineId, "worker_id", candidate.NodeId)

	member, err := m.clusterState.GetNode(context.Background(), candidate.NodeId)
	if err != nil {
		slog.Error("Failed to get node", "node_id", candidate.NodeId)
		return ravelerrors.NewUnknown("Unknown error during placement")
	}

	conn, err := grpc.NewClient(member.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to agent", "node_id", candidate.NodeId)
		return ravelerrors.NewUnavailable("Failed to connect to agent")
	}

	agentClient := proto.NewAgentServiceClient(conn)

	resp, err := agentClient.CreateInstance(context.Background(), &proto.CreateInstanceRequest{
		MachineId: machineId,
		Namespace: namespace,
		FleetId:   fleetId,
		Config:    core.MachineConfigToProto(options.Config),
		Start:     !options.SkipStart,
	})
	if err != nil {
		slog.Error("Failed to create instance for machine", "error", err)
		return ravelerrors.NewUnknown("Failed to create machine")
	}

	if err := m.clusterState.CreateMachine(ctx, cluster.Machine{
		Id:         machineId,
		Namespace:  namespace,
		FleetId:    fleetId,
		InstanceId: resp.InstanceId,
		Region:     options.Region,
		Node:       candidate.NodeId,
		CreatedAd:  createdAt,
		UpdatedAt:  createdAt,
		Destroyed:  false,
	}); err != nil {
		slog.Error("Failed to create machine in corrosion", "machine_id", machineId, "error", err)
		return err
	}

	return nil
}
