package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/internal/agent/instance"
	"github.com/valyentdev/ravel/internal/agent/instance/state"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/proto"
)

func (d *Agent) CreateInstance(ctx context.Context, opt *proto.CreateInstanceRequest) (*proto.CreateInstanceResponse, error) {
	err := d.reservations.ConfirmReservation(ctx, opt.MachineId)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm reservation: %w", err)
	}

	config := core.MachineConfigFromProto(opt.Config)

	var desiredStatus core.InstanceStatus

	if opt.Start {
		desiredStatus = core.MachineStatusRunning
	} else {
		desiredStatus = core.MachineStatusStopped
	}

	i := core.Instance{
		Id:            id.GeneratePrefixed("instance"),
		Namespace:     opt.Namespace,
		MachineId:     opt.MachineId,
		FleetId:       opt.FleetId,
		DesiredStatus: desiredStatus,
		Config:        config,
		CreatedAt:     time.Now(),
	}

	i.Config = config
	i.NodeId = d.nodeId

	s := state.NewInstanceState(d.store, i, nil, d.nodeId, d.cluster)

	err = s.Create(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	manager := instance.NewInstanceManager(s, d.containerRuntime)
	d.lock.Lock()
	d.instances[i.Id] = manager
	d.lock.Unlock()

	go func() {
		err := manager.Prepare()
		if err != nil {
			slog.Error("failed to prepare instance", "instance", i.Id, "error", err)
			return
		}
	}()

	return &proto.CreateInstanceResponse{
		InstanceId: i.Id,
	}, nil
}
