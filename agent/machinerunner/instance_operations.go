package machinerunner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func errMachineIs(status api.MachineStatus) error {
	return errdefs.NewFailedPrecondition(fmt.Sprintf("machine is in %s status", status))
}

func (m *MachineRunner) canUseInstance() error {
	status := m.state.Status()
	if status == api.MachineStatusDestroyed || status == api.MachineStatusDestroying || status == api.MachineStatusPreparing || status == api.MachineStatusCreated {
		return errMachineIs(status)
	}
	return nil
}

func (m *MachineRunner) Start(ctx context.Context) error {
	prev, _, err := m.state.PushStartEvent(false)
	if err != nil {
		if errdefs.IsFailedPrecondition(err) && (prev.Status == api.MachineStatusStarting || prev.Status == api.MachineStatusRunning) {
			return nil
		}
		return err
	}

	return nil
}

func (m *MachineRunner) startInstance() {
	go func() {
		m.runLock.Lock()
		defer m.runLock.Unlock()
		ctx := context.Background()

		instanceId := m.state.InstanceId()
		i, err := m.runtime.GetInstance(instanceId)
		if err != nil {
			m.state.PushStartFailedEvent(err.Error())
			return
		}

		if i.State.Status == instance.InstanceStatusRunning || i.State.Status == instance.InstanceStatusStarting {
			m.state.PushStartedEvent()
			return
		}

		err = m.runtime.StartInstance(ctx, m.state.InstanceId())
		if err != nil {
			m.state.PushStartFailedEvent(err.Error())
			return
		}
		m.state.PushStartedEvent()
	}()
}

func (m *MachineRunner) Stop(ctx context.Context, stopConfig *api.StopConfig) error {
	slog.Info("Stopping machine", "machine_id", m.state.Id())
	prev, _, err := m.state.PushStopEvent(api.MachineStopEventPayload{
		Config: stopConfig,
	})
	if err != nil {
		if errdefs.IsFailedPrecondition(err) && prev.Status == api.MachineStatusStopped {
			return nil
		}
		return err
	}

	return nil
}

func (m *MachineRunner) stopInstance(ctx context.Context, stopConfig *api.StopConfig) error {
	err := m.runtime.StopInstance(ctx, m.state.InstanceId(), stopConfig)
	if err != nil {
		m.state.PushStopFailedEvent()
		return nil
	}

	return nil
}

func (m *MachineRunner) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	if err := m.canUseInstance(); err != nil {
		return nil, err
	}

	status := m.state.Status()
	if status != api.MachineStatusRunning {
		return nil, errMachineIs(status)
	}

	return m.runtime.InstanceExec(ctx, m.state.InstanceId(), cmd, timeout)
}

func (m *MachineRunner) GetLogs() ([]*api.LogEntry, error) {
	if err := m.canUseInstance(); err != nil {
		return nil, err
	}
	return m.runtime.GetInstanceLogs(m.state.InstanceId())
}

func (m *MachineRunner) SubscribeToLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	if err := m.canUseInstance(); err != nil {
		return nil, nil, err
	}
	return m.runtime.SubscribeToInstanceLogs(ctx, m.state.MachineInstance().Machine.InstanceId)
}
