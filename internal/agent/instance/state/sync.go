package state

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/internal/cluster"
)

func (i *instanceState) sync() {
	for range i.updateCh {
		slog.Info("syncing instance state", "instance", i.instance.Id, "conf", i.instance.Config)
		instance := i.instance
		event := i.lastEvent
		err := i.cluster.UpsertInstance(context.Background(), cluster.Instance{
			Id:        instance.Id,
			MachineId: instance.MachineId,
			Node:      i.node,
			Config:    instance.Config,
			ImageRef:  instance.Config.Workload.Image,
			CreatedAt: instance.CreatedAt,
			UpdatedAt: event.Timestamp,
		}, cluster.InstanceStatus{
			Id:        instance.Id,
			MachineId: instance.MachineId,
			Status:    event.Status,
		})

		if err != nil {
			slog.Error("failed to upsert instance", "err", err)
			go func() {
				time.Sleep(5 * time.Second)
				i.updateCh <- struct{}{}
			}()
		}

	}

}
