package instance

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valyentdev/ravel/internal/agent/tap"
	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

func (m *Manager) Start(ctx context.Context) error {
	if !m.mutex.TryLock() {
		status := m.state.Status()
		// if the status is running or starting, we return nil as the instance is already running
		if status == core.MachineStatusRunning || status == core.MachineStatusStarting {
			return nil
		}
		// if the status is stopped, we try to get the lock to start the instance if possible after getting the lock
		if status == core.MachineStatusStopped {
			m.mutex.Lock()
		} else {
			// in all other cases the instance cannot be started now and another incompatible operation is probably in progress
			return core.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", status))
		}
	}
	defer m.mutex.Unlock()

	if !canTransition(m.state.Status(), core.MachineStatusStarting) {
		return core.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", m.state.Status()))
	}

	if m.isRunning {
		return nil
	}

	err := m.state.PushInstanceStartEvent(ctx, false)
	if err != nil {
		return err
	}

	netConfig := networking.LocalIPV4Subnet(m.reservation.LocalIPV4Subnet).LocalConfig()

	_, err = tap.PrepareInstanceTapDevice(m.Instance(), netConfig)
	if err != nil {
		if err := m.state.PushInstanceStartFailedEvent(ctx, err.Error()); err != nil {
			slog.Error("failed to push instance failed event", "error", err)
		}
		return err
	}

	h, err := m.runtime.StartVM(context.Background(), m.state.Instance())
	if err != nil {
		if err := m.state.PushInstanceStartFailedEvent(ctx, err.Error()); err != nil {
			slog.Error("failed to push instance failed event", "error", err)
		}
		return err
	}

	err = m.state.PushInstanceStartedEvent(ctx)
	if err != nil {
		slog.Error("failed to push instance started event", "error", err)
	}

	m.isRunning = true
	m.waitCh = make(chan struct{})

	go m.run(h)

	return nil
}

func (m *Manager) run(h runtimes.Handle) {
	if h.Console != "" {
		err := m.logger.Start(h.Console)
		if err != nil {
			slog.Error("failed to start logger", "error", err)
		}

		defer m.logger.Stop()
	}

	result, err := m.runtime.WaitVM(context.Background(), m.state.Instance().Id)
	if err != nil {
		slog.Error("failed to wait vm", "error", err)
		return
	}

	err = m.runtime.DestroyVM(context.Background(), m.state.Instance().Id)
	if err != nil {
		slog.Error("failed to destroy vm", "error", err)
	}

	err = m.state.PushInstanceExitedEvent(context.Background(), *result)
	if err != nil {
		slog.Error("failed to push instance exited event", "error", err)
	}

	err = tap.CleanupInstanceTapDevice(m.Instance().Id, networking.LocalIPV4Subnet(m.reservation.LocalIPV4Subnet).LocalConfig())
	if err != nil {
		slog.Error("failed to cleanup tap device", "error", err)
	}

	m.isRunning = false
	close(m.waitCh)

}
