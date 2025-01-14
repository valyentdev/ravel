package machinerunner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
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

func (m *MachineRunner) lockForInstance() error {
	if !m.mutex.TryLock() {
		if err := m.canUseInstance(); err != nil {
			return err
		}

		m.mutex.Lock()
		return nil
	}
	return nil
}

func (m *MachineRunner) unlockForInstance() {
	m.mutex.Unlock()
}

func (m *MachineRunner) Start(ctx context.Context) error {
	err := m.lockForInstance()
	if err != nil {
		return err
	}
	defer m.unlockForInstance()

	if m.state.Status() == api.MachineStatusStarting || m.state.Status() == api.MachineStatusRunning {
		return nil
	}

	if m.state.Status() != api.MachineStatusStopped {
		return errMachineIs(m.state.Status())
	}

	m.start(false)
	return nil
}

func (m *MachineRunner) start(isRestart bool) {
	slog.Info("Starting machine", "machine_id", m.state.InstanceId())
	ctx := context.Background()
	m.state.PushStartEvent(isRestart)
	go func() {
		m.runLock.Lock()
		defer m.runLock.Unlock()
		err := m.runtime.StartInstance(ctx, m.state.InstanceId())
		if err != nil {
			m.state.PushStartFailedEvent(err.Error())
			return
		}
		m.state.PushStartedEvent()
	}()
}

func (m *MachineRunner) Stop(ctx context.Context, stopConfig *api.StopConfig) error {
	err := m.lockForInstance()
	if err != nil {
		return err
	}
	defer m.unlockForInstance()

	status := m.state.Status()

	if status == api.MachineStatusStopped || status == api.MachineStatusStopping {
		return nil
	}
	if status != api.MachineStatusRunning {
		return errMachineIs(status)
	}

	return m.stop(ctx, stopConfig)
}

func (m *MachineRunner) stop(ctx context.Context, stopConfig *api.StopConfig) error {
	err := m.state.PushStopEvent(api.MachineStopEventPayload{
		Config: stopConfig,
	})
	if err != nil {
		return err
	}

	go func() {
		m.runLock.Lock()
		defer m.runLock.Unlock()
		err := m.runtime.StopInstance(ctx, m.state.InstanceId(), stopConfig)
		if err != nil {
			slog.Error("Failed to stop instance", "error", err)
			if !errdefs.IsFailedPrecondition(err) {
				m.state.PushStopFailedEvent()
			}
		}
	}()

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
