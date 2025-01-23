package agent

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/agent/machinerunner/state"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/internal/eventer"
)

func newMachineEventer(store state.Store, nc *nats.Conn) *eventer.Eventer[api.MachineEvent] {
	return eventer.NewEventer(eventer.Options[api.MachineEvent]{
		Report: func(e *api.MachineEvent) error {
			bytes, err := json.Marshal(e)
			if err != nil {
				return err
			}

			_, err = nc.Request("machines.events", bytes, time.Second)
			if err != nil {
				return err
			}

			return nil
		},
		OnSuccess: func(e *api.MachineEvent) {
			err := store.DeleteMachineInstanceEvent(e.Id)
			if err != nil {
				slog.Error("failed to delete event", "error", err)
			}
		},
		OnError: func(e *api.MachineEvent) bool {
			slog.Error("failed to report machine event", "event", e.Id)
			return true
		},
		Backoff: time.Second,
	})

}
