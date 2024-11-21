package instancemanager

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *Manager) Stop(ctx context.Context, signal string, timeout time.Duration) error {
	m.state.Lock()
	status := m.state.Status()

	if status == instance.InstanceStatusStopped {
		m.state.Unlock()
		return nil
	}

	if status != instance.InstanceStatusRunning && status != instance.InstanceStatusStopping {
		m.state.Unlock()
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", status))
	}

	if err := m.state.UpdateInstanceState(instance.State{
		Status: instance.InstanceStatusStopping,
	}); err != nil {
		m.state.Unlock()
		return err
	}

	// We need to unlock the state before calling stopImpl to allow user to send another stop request with different signal or timeout
	m.state.Unlock()

	return m.stopImpl(ctx, signal, timeout)
}

func (m *Manager) stopImpl(ctx context.Context, signal string, timeout time.Duration) error {
	vm := m.getVM()

	err := vm.Stop(context.Background(), signal)
	if err != nil {
		slog.Error("failed to stop vm", "error", err)
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	exited := vm.WaitExit(ctxTimeout)
	if exited {
		m.waitExit()
		return nil
	}

	err = vm.Shutdown(context.Background())
	if err != nil {
		return err
	}

	m.waitExit()
	return nil
}
