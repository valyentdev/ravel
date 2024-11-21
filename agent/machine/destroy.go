package machine

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (m *Machine) Destroy(ctx context.Context, force bool) error {
	if m.destroying.Load() || m.state.Status() == api.MachineStatusDestroyed {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := m.state.Status()

	if status == api.MachineStatusStopped {
		go m.destroyImpl(ctx)
		return nil
	}

	if !force {
		return errdefs.NewFailedPrecondition("machine is not stopped")
	}

	timeout := 0

	if err := m.runtime.StopInstance(ctx, m.state.InstanceId(), &api.StopConfig{
		Timeout: &timeout,
	}); err != nil {
		return err
	}

	m.destroying.Store(true)
	go m.destroyImpl(ctx)
	return nil

}

func (m *Machine) destroyImpl(ctx context.Context) {
	m.destroying.Store(true)
	err := m.runtime.DestroyInstance(ctx, m.state.InstanceId())
	if err != nil && !errdefs.IsNotFound(err) {
		slog.Error("failed to destroy instance", "instance", m.state.InstanceId(), "error", err)
		return
	}

	err = m.state.PushDestroyedEvent(ctx)
	if err != nil {
		slog.Error("failed to push destroyed event", "instance", m.state.InstanceId(), "error", err)
	}

	m.onDestroyed(m.state.MachineInstance())
}
