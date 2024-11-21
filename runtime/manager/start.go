package instancemanager

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *Manager) Start(ctx context.Context) error {
	m.state.Lock()
	defer m.state.Unlock()

	slog.Debug("starting instance", "id", m.state.Instance().Id)

	if m.state.Status() != instance.InstanceStatusStopped {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", m.state.Status()))
	}

	err := m.state.UpdateInstanceState(instance.State{
		Status: instance.InstanceStatusStarting,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			m.onStartFailed()
		}
	}()

	i := m.state.Instance()

	err = m.networking.EnsureInstanceNetwork(i.Id, i.Network)
	if err != nil {
		slog.Error("failed to ensure instance network", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			err := m.networking.CleanupInstanceNetwork(i.Id, i.Network)
			if err != nil {
				slog.Error("failed to cleanup instance network", "error", err)
			}
		}
	}()

	vm, err := m.vmBuilder.BuildInstanceVM(ctx, &i)
	if err != nil {
		slog.Error("failed to build vm", "error", err)
		return err
	}
	defer func() {
		if err != nil {
			err := m.vmBuilder.CleanupInstanceVM(ctx, &i)
			if err != nil {
				slog.Error("failed to destroy vm", "error", err)
			}
		}
	}()

	m.setVM(vm)

	h, err := vm.Start(ctx)
	if err != nil {
		return err
	}

	m.waitCh = make(chan struct{})

	err = m.state.UpdateInstanceState(instance.State{
		Status: instance.InstanceStatusRunning,
	})
	if err != nil {
		slog.Error("failed to update instance state", "error", err)
		err = nil // We continue anyway, things are already started
	}

	m.er.ReportInstanceEvent(instance.Event{
		Event:            instance.InstanceStarted,
		InstanceId:       m.state.Instance().Id,
		InstanceMetadata: m.state.Instance().Metadata,
		Timestamp:        time.Now(),
	})

	go m.run(h)

	return err
}

func (m *Manager) onStartFailed() {
	m.state.UpdateInstanceState(instance.State{
		Status: instance.InstanceStatusStopped,
	})
}
