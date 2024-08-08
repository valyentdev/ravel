package instance

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
)

func (m *Manager) Recover() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	slog.Info("recovering instance", "instance_id", m.state.Id(), "status", m.state.Status())
	if m.state.Status() == core.MachineStatusRunning || m.state.Status() == core.MachineStatusStarting || m.state.Status() == core.MachineStatusStopping {
		m.recoverRunning(m.state.Status() == core.MachineStatusStarting)
		return
	}

}

func (m *Manager) recoverRunning(wasStarting bool) {
	slog.Info("recovering running instance", "instance_id", m.state.Id())
	stillRunning := m.runtime.RecoverVM(context.Background(), m.state.Instance())

	if stillRunning {
		go m.run()
		return
	}

	if wasStarting {
		err := m.state.PushInstanceStartFailedEvent(context.Background(), "A start event was in progress when it was cancelled by ravel or host failure")
		if err != nil {
			slog.Error("failed to push instance start failed event", "error", err)
		}
		return
	}

	err := m.state.PushInstanceExitedEvent(context.Background(), core.InstanceExitedEventPayload{
		Success:   false,
		Requested: false,
		ExitedAt:  time.Now(),
	})

	if err != nil {
		slog.Error("failed to push instance exited event", "error", err)
	}

}
