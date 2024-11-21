package machine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

var ErrInstanceIsDestroying = errdefs.NewFailedPrecondition("instance is destroying")

func (m *Machine) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.state.Status() == api.MachineStatusRunning || m.state.Status() == api.MachineStatusStarting {
		return nil
	}

	if m.state.Status() != api.MachineStatusStopped {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("machine is in %s status", m.state.Status()))
	}

	err := m.state.PushStartEvent(false)
	if err != nil {
		return err
	}

	go func() {
		err := m.runtime.StartInstance(context.Background(), m.state.InstanceId())
		if err != nil {
			if err := m.state.PushStartFailedEvent(context.Background(), err.Error()); err != nil {
				slog.Error("Failed to push start failed event", "error", err)
			}
		}
	}()

	return nil
}

func (m *Machine) Stop(ctx context.Context, stopConfig *api.StopConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	status := m.state.Status()

	if status == api.MachineStatusStopped || status == api.MachineStatusStopping {
		return nil
	}
	if status != api.MachineStatusRunning {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("machine is in %s status", m.state.Status()))
	}

	return m.stop(ctx, stopConfig)
}

func (m *Machine) stop(ctx context.Context, stopConfig *api.StopConfig) error {
	err := m.state.PushStopEvent()
	if err != nil {
		return err
	}

	go func() {
		err := m.runtime.StopInstance(ctx, m.state.InstanceId(), stopConfig)
		if err != nil {
			slog.Error("Failed to stop instance", "error", err)
		}
	}()

	return nil
}

func (m *Machine) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	return m.runtime.InstanceExec(ctx, m.state.InstanceId(), cmd, timeout)
}

func (m *Machine) GetLogs() ([]*api.LogEntry, error) {
	return m.runtime.GetInstanceLogs(m.state.InstanceId())
}

func (m *Machine) SubscribeToLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	return m.runtime.SubscribeToInstanceLogs(ctx, m.state.MachineInstance().Machine.InstanceId)
}
