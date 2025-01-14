package machinerunner

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

func (m *MachineRunner) Destroy(ctx context.Context, force bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.state.Status() == api.MachineStatusDestroying || m.state.Status() == api.MachineStatusDestroyed {
		return nil
	}

	status := m.state.Status()

	if status == api.MachineStatusStopped {
		go m.destroyImpl(ctx, force, "requested by user")
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

	go m.destroyImpl(ctx, force, "requested by user")
	return nil

}

func (m *MachineRunner) destroyImpl(ctx context.Context, force bool, reason string) error {
	if err := m.state.PushDestroyEvent(api.OriginUser, force, reason); err != nil {
		return err
	}
	err := m.runtime.DestroyInstance(ctx, m.state.InstanceId())
	if err != nil && !errdefs.IsNotFound(err) {
		slog.Error("failed to destroy instance", "instance", m.state.InstanceId(), "error", err)
		return err
	}

	err = m.state.PushDestroyedEvent(ctx)
	if err != nil {
		return err
	}

	m.onDestroyed(m.state.MachineInstance())
	return nil
}
