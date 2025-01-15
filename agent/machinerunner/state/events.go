package state

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
)

const maxSize = 10

func (is *MachineInstanceState) persistStateAndEvent(e api.MachineEvent) error {
	previousEvents := is.mi.State.LastEvents
	currentSize := len(previousEvents)

	lastIndex := currentSize - 1

	newSize := currentSize + 1
	if currentSize == maxSize {
		previousEvents = previousEvents[:lastIndex-1]
		newSize = maxSize
	}

	newEvents := make([]api.MachineEvent, 0, newSize)

	newEvents = append(newEvents, e)
	newEvents = append(newEvents, previousEvents...)

	is.mi.State.LastEvents = newEvents

	if err := is.store.PutMachineInstanceEvent(e); err != nil {
		return err
	}

	is.statusObserver.Set(e.Status)

	err := is.persistState()
	is.eventer.ReportEvent(e)

	return err
}

func (s *MachineInstanceState) newEvent(eventType api.MachineEventType, origin api.Origin, status api.MachineStatus, payload api.MachineEventPayload) api.MachineEvent {
	return api.MachineEvent{
		Id:         eventId(),
		Type:       eventType,
		Origin:     origin,
		Payload:    payload,
		MachineId:  s.mi.Machine.Id,
		InstanceId: s.mi.Machine.InstanceId,
		Status:     status,
		Timestamp:  time.Now(),
	}
}

func (s *MachineInstanceState) LastEvent() api.MachineEvent {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.mi.State.LastEvents) == 0 {
		return api.MachineEvent{}
	}

	return s.mi.State.LastEvents[0]
}

func (s *MachineInstanceState) PushPrepareEvent(ctx context.Context) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachinePrepare,
		api.OriginRavel,
		api.MachineStatusPreparing,
		api.MachineEventPayload{},
	)

	s.mi.State.Status = api.MachineStatusPreparing

	err = s.persistStateAndEvent(event)
	if err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushPreparedEvent(ctx context.Context, network instance.NetworkingConfig) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachinePrepared,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{},
	)

	s.mi.State.Status = api.MachineStatusStopped
	s.mi.State.LocalIPV4 = network.Local.InstanceIP.String()

	err = s.persistStateAndEvent(event)
	if err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushPrepareFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachinePrepareFailed,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{
			PrepareFailed: &api.MachinePrepareFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	s.mi.State.Status = api.MachineStatusDestroying
	s.mi.State.DesiredStatus = api.MachineStatusDestroyed

	err = s.persistStateAndEvent(event)

	return nil
}

func (s *MachineInstanceState) PushStartEvent(isRestart bool) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var origin api.Origin
	if isRestart {
		origin = api.OriginRavel
	} else {
		origin = api.OriginUser
	}

	event := s.newEvent(
		api.MachineStart,
		origin,
		api.MachineStatusStarting,
		api.MachineEventPayload{
			Start: &api.MachineStartEventPayload{
				IsRestart: isRestart,
			},
		},
	)

	s.mi.State.DesiredStatus = api.MachineStatusRunning
	s.mi.State.Status = api.MachineStatusStarting
	if isRestart {
		s.mi.State.Restarts = s.mi.State.Restarts + 1
	} else {
		s.mi.State.Restarts = 0
	}

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushStartFailedEvent(errMsg string) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineStartFailed,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{
			StartFailed: &api.MachineStartFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	s.mi.State.Status = api.MachineStatusStopped

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushStartedEvent() (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineStarted,
		api.OriginRavel,
		api.MachineStatusRunning,
		api.MachineEventPayload{
			Started: &api.MachineStartedEventPayload{
				StartedAt: time.Now(),
			},
		},
	)

	s.mi.State.Status = api.MachineStatusRunning

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushStopEvent(payload api.MachineStopEventPayload) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineStop,
		api.OriginUser,
		api.MachineStatusStopping,
		api.MachineEventPayload{},
	)

	s.mi.State.Status = api.MachineStatusStopping
	s.mi.State.DesiredStatus = api.MachineStatusStopped
	s.mi.State.Restarts = 0

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}
	return nil
}

func (s *MachineInstanceState) PushStopFailedEvent() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineStopFailed,
		api.OriginRavel,
		api.MachineStatusRunning,
		api.MachineEventPayload{},
	)

	s.mi.State.Status = api.MachineStatusRunning

	if err := s.persistStateAndEvent(event); err != nil {
		return err
	}

	return nil
}

func (s *MachineInstanceState) PushExitedEvent(payload api.MachineExitedEventPayload) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineExited,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{
			Exited: &payload,
		},
	)

	s.mi.State.Status = api.MachineStatusStopped

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushDestroyEvent(origin api.Origin, force bool, autoDestroy bool, reason string) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineDestroy,
		origin,
		api.MachineStatusDestroying,
		api.MachineEventPayload{
			Destroy: &api.MachineDestroyEventPayload{
				Reason:      reason,
				Force:       force,
				AutoDestroy: autoDestroy,
			},
		},
	)

	s.mi.State.Status = api.MachineStatusDestroying
	s.mi.State.DesiredStatus = api.MachineStatusDestroyed

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushDestroyedEvent(ctx context.Context) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineDestroyed,
		api.OriginRavel,
		api.MachineStatusDestroyed,
		api.MachineEventPayload{},
	)

	s.mi.State.Status = api.MachineStatusDestroyed
	s.mi.State.DesiredStatus = api.MachineStatusDestroyed

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}

func (s *MachineInstanceState) PushGatewayEvent(enabled bool) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	event := s.newEvent(
		api.MachineGateway,
		api.OriginUser,
		s.mi.State.Status,
		api.MachineEventPayload{
			Gateway: &api.MachineGatewayEventPayload{
				Enabled: enabled,
			},
		},
	)

	s.mi.State.MachineGatewayEnabled = enabled

	if err = s.persistStateAndEvent(event); err != nil {
		return
	}

	return nil
}
