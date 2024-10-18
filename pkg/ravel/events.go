package ravel

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/pkg/core"
)

func (r *Ravel) ListenInstanceEvents() error {
	_, err := r.nc.QueueSubscribe("events.*.*", "servers", func(msg *nats.Msg) {
		var event core.InstanceEvent
		err := json.Unmarshal(msg.Data, &event)
		if err != nil {
			slog.Info("failed to unmarshal message", "error", err)
			return
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
