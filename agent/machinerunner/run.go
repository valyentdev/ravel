package machinerunner

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *MachineRunner) Run() {
	m.mutex.Lock()

	shouldContinue := m.recover(context.Background())
	if !shouldContinue {
		m.mutex.Unlock()
		return
	}

	ctx := context.Background()
	updates, err := m.runtime.WatchInstanceState(ctx, m.state.InstanceId())
	if err != nil {
		if errdefs.IsNotFound(err) {
			m.destroyImpl(ctx, destroyPayload{
				origin: api.OriginRavel,
				reason: "instance not found",
				force:  true,
			})
		}

		slog.Error("failed to watch instance state", "machine_id", m.state.Id(), "error", err)
		return
	}
	m.mutex.Unlock()
	for {
		select {
		case instanceUpdate := <-updates:
			m.runLock.Lock()
			m.onInstanceUpdate(&instanceUpdate)
			m.runLock.Unlock()
		case <-ctx.Done():
			return
		}

	}
}

func (m *MachineRunner) recover(ctx context.Context) bool {
	status := m.state.Status()

	switch status {
	case api.MachineStatusCreated, api.MachineStatusPreparing:
		err := m.prepare(ctx)
		if err != nil {
			return false
		}
	case api.MachineStatusDestroyed:
		m.onDestroyed(m.state.MachineInstance())
		return false

	case api.MachineStatusDestroying:
		err := m.destroyImpl(ctx, destroyPayload{
			origin: api.OriginRavel,
			reason: "recover from destroying",
			force:  true,
		})
		if err != nil {
			slog.Error("failed to destroy machine", "machine_id", m.state.Id(), "error", err)
		}
		return false
	}

	return true

}

func (m *MachineRunner) onInstanceUpdate(update *instance.State) {
	state := m.state.State()
	status := state.Status
	instanceStatus := update.Status

	switch instanceStatus {
	case instance.InstanceStatusCreated:
		if instanceStatus == instance.InstanceStatusCreated {
			if state.DesiredStatus == api.MachineStatusRunning {
				m.start(false)
			}
		}
	case instance.InstanceStatusStopped:
		if status == api.MachineStatusStarting {
			m.state.PushStartFailedEvent("instance not running")
		}
		if status == api.MachineStatusRunning || status == api.MachineStatusStopping {
			var payload api.MachineExitedEventPayload
			if update.ExitResult != nil {
				payload.ExitCode = update.ExitResult.ExitCode
				payload.ExitedAt = update.ExitResult.ExitedAt
			} else {
				payload.ExitCode = -1
				payload.ExitedAt = time.Now()
			}

			m.state.PushExitedEvent(payload)
			m.handleExit(&payload)
			return
		}
	case instance.InstanceStatusRunning:
		if status == api.MachineStatusStopped {
			m.state.PushStartedEvent()
		}
	}
}
