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
)

func (d *Agent) CreateInstance(ctx context.Context, opt core.CreateInstancePayload) (*core.Instance, error) {
	reservation, err := d.reservations.ConfirmReservation(ctx, opt.MachineId)
	if err != nil {
		return nil, fmt.Errorf("failed to confirm reservation: %w", err)
	}

	var desiredStatus core.InstanceStatus

	if opt.Start {
		desiredStatus = core.MachineStatusRunning
	} else {
		desiredStatus = core.MachineStatusStopped
	}

	config := opt.Config
	i := core.Instance{
		Id:        id.GeneratePrefixed("instance"),
		Namespace: opt.Namespace,
		MachineId: opt.MachineId,
		FleetId:   opt.FleetId,
		State: core.InstanceState{
			DesiredStatus: desiredStatus,
			Status:        core.MachineStatusCreated,
		},
		Config:    config,
		CreatedAt: time.Now(),
		LocalIPV4: reservation.LocalIPV4Subnet.LocalConfig().MachineIP.String(),
	}

	i.Config = config
	i.NodeId = d.nodeId

	is, err := state.CreateInstance(d.store, d.cluster, i)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	manager := instance.NewInstanceManager(is, d.containerRuntime, reservation)
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

	return &i, nil
}
