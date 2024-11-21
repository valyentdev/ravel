package ravel

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/api"
)

func (r *Ravel) ListenInstanceEvents() error {
	_, err := r.nc.QueueSubscribe("machines.events", "servers", func(msg *nats.Msg) {
		var event api.MachineEvent
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			slog.Info("failed to unmarshal message", "error", err)
			return
		}

		if event.Type == api.MachineDestroyed {
			err = r.state.DestroyMachine(context.Background(), event.MachineId)
			if err != nil {
				slog.Info("failed to destroy machine", "error", err)
				return
			}
		}

		err = r.db.PushMachineEvent(context.Background(), event)
		if err != nil {
			slog.Info("failed to push machine event", "error", err)
			return
		}

		err = msg.Respond([]byte("ok"))
		if err != nil {
			slog.Info("failed to respond to message", "error", err)
		}
	})

	if err != nil {
		return err
	}

	return nil
}
