package machine

import (
	"log/slog"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *Machine) ProcessInstanceEvent(event instance.Event) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var err error
	switch event.Event {
	case instance.InstanceExited:
		err = m.state.PushExitedEvent(api.MachineExitedEventPayload{
			ExitedAt: event.Payload.Exited.ExitedAt,
			ExitCode: int(event.Payload.Exited.ExitCode),
		})
	case instance.InstanceStarted:
		err = m.state.PushStartedEvent()
	}
	if err != nil {
		slog.Error("failed to process machine instance event", "error", err)
	}

}
