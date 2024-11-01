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

	if m.state.Status() == core.MachineStatusPreparing {
		m.recoverPreparing()
		return
	}

}

func (m *Manager) recoverRunning(wasStarting bool) {
	slog.Info("recovering running instance", "instance_id", m.state.Id())
	h, stillRunning := m.runtime.RecoverVM(context.Background(), m.state.Instance())

	if stillRunning {
		m.waitCh = make(chan struct{})
		m.isRunning = true
		go m.run(h)
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

func (m *Manager) recoverPreparing() {
	slog.Info("recovering preparing instance", "instance_id", m.state.Id())
	err := m.prepare()
	if err != nil {
		slog.Error("failed to recover preparing instance", "instance_id", m.state.Id(), "error", err)
	}
}
