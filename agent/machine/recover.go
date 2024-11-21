package machine

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
)

func (m *Machine) Recover() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	slog.Info("Recovering", "machine_id", m.state.Id())
	status := m.state.Status()
	switch status {
	case api.MachineStatusCreated, api.MachineStatusPreparing:
		m.recoverPrepare()
		return
	case api.MachineStatusDestroyed:
		m.onDestroyed(m.state.MachineInstance())
		return
	case api.MachineStatusDestroying:
		m.Destroy(context.Background(), true)
		return
	case api.MachineStatusStarting:
		m.recoverStopping()
	case api.MachineStatusStopping:
		m.recoverStopping()
	}
}

// lock is held
func (m *Machine) recoverPrepare() {
	err := m.prepare(context.Background())
	if err != nil {
		return
	}
}

// lock is held
func (m *Machine) recoverStopping() {
	err := m.stop(context.Background(), nil)
	if err != nil {
		slog.Error("failed to stop machine", "machine_id", m.state.Id(), "error", err)
		return
	}
}
