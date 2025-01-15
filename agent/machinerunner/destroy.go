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
	payload := destroyPayload{
		origin:      api.OriginUser,
		reason:      "requested by user",
		autoDestroy: false,
		force:       force,
	}

	status := m.state.Status()

	if status == api.MachineStatusStopped {
		go m.destroyImpl(ctx, payload)
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

	go m.destroyImpl(ctx, payload)
	return nil

}

type destroyPayload struct {
	origin      api.Origin
	reason      string
	autoDestroy bool
	force       bool
}

func (m *MachineRunner) destroyImpl(ctx context.Context, p destroyPayload) error {
	origin := api.OriginUser
	if p.origin != "" {
		origin = p.origin
	}

	if err := m.state.PushDestroyEvent(origin, p.force, p.autoDestroy, p.reason); err != nil {
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
