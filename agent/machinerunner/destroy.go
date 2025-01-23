package machinerunner

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

func (m *MachineRunner) Destroy(ctx context.Context, force bool) error {
	prev, _, err := m.state.PushDestroyEvent(api.OriginUser, force, false, "requested by user")
	if err != nil {
		if errdefs.IsFailedPrecondition(err) && (prev.Status == api.MachineStatusDestroyed || prev.Status == api.MachineStatusDestroying) {
			return nil
		}
		return err
	}
	return nil
}

type destroyPayload struct {
	origin      api.Origin
	reason      string
	autoDestroy bool
	force       bool
}

func (m *MachineRunner) destroyImpl() {
	err := m.stopAndDestroyInstance()
	if err != nil {
		slog.Error("failed to stop and destroy instance", "instance", m.state.InstanceId(), "error", err)
		return
	}

	_, _, err = m.state.PushDestroyedEvent()
	if err != nil {
		slog.Error("failed to push destroyed event", "instance", m.state.InstanceId(), "error", err)
	}

}

func (m *MachineRunner) stopAndDestroyInstance() error {
	timeout := 0
	if err := m.runtime.StopInstance(context.Background(), m.state.InstanceId(), &api.StopConfig{
		Timeout: &timeout,
	}); err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}

		return err
	}

	if err := m.runtime.DestroyInstance(context.Background(), m.state.InstanceId()); err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}

		return err
	}

	return nil
}
