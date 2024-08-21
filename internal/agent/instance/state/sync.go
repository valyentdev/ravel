package state

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/internal/cluster"
)

func (i *instanceState) sync() {
	for range i.updateCh {
		instance := i.instance
		event := i.lastEvent
		err := i.cluster.UpsertInstance(context.Background(), cluster.Instance{
			Id:             instance.Id,
			Node:           i.node,
			MachineId:      instance.MachineId,
			MachineVersion: instance.MachineVersion,
			Status:         event.Status,
			CreatedAt:      instance.CreatedAt,
			UpdatedAt:      event.Timestamp,
		})

		if err != nil {
			slog.Error("failed to upsert instance", "err", err)
			go func() {
				time.Sleep(5 * time.Second)
				i.triggerUpdate()
			}()
		}

	}

}
