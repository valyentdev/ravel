package instance

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

func (m *Manager) Stop(ctx context.Context, signal string, timeout time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	currentStatus := m.state.Status()

	if !canTransition(m.state.Status(), core.MachineStatusStopping) {
		if !m.isRunning {
			return nil
		}
		// We allow to send a second stop request even if the instance is already stopping (e.g. to change the signal or timeout)
		if currentStatus != core.MachineStatusStopping {
			return ravelerrors.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", currentStatus))
		}
	}

	err := m.state.PushInstanceStopEvent(ctx)
	if err != nil {
		return err
	}

	go func() {
		err := m.runtime.StopVM(context.Background(), m.state.Id(), signal, timeout)
		if err != nil {
			slog.Error("failed to stop vm", "error", err)
		}
	}()

	return nil
}
