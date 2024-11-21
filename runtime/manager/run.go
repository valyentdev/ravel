package instancemanager

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/core/instance"
)

func (m *Manager) run(h instance.Handle) {
	vmm := m.getVM()
	if h.Console != "" {
		err := m.logger.Start(h.Console)
		if err != nil {
			slog.Error("failed to start logger", "error", err)
		}

		defer m.logger.Stop()
	}

	result := vmm.Run()

	m.postRunHook()

	err := m.state.UpdateInstanceState(instance.State{
		Status:     instance.InstanceStatusStopped,
		ExitResult: result,
	})
	if err != nil {
		slog.Error("failed to update instance state", "error", err)
	}

	m.er.ReportInstanceEvent(instance.Event{
		Event:            instance.InstanceExited,
		InstanceId:       m.Instance().Id,
		InstanceMetadata: m.Instance().Metadata,
		Payload: instance.InstanceEventPayload{
			Exited: result,
		},
		Timestamp: time.Now(),
	})

	close(m.waitCh)
}

func (m *Manager) postRunHook() {
	i := m.Instance()
	err := m.vmBuilder.CleanupInstanceVM(context.Background(), &i)
	if err != nil {
		slog.Error("failed to destroy vm", "error", err)
	}

	err = m.networking.CleanupInstanceNetwork(i.Id, i.Network)
	if err != nil {
		slog.Error("failed to cleanup instance network", "error", err)
	}
}
