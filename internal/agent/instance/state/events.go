package state

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
)

func (is *instanceState) persistChange(event core.InstanceEvent) error {
	is.lastEvent = event
	is.instance.State.Status = event.Status

	err := is.store.UpdateInstance(is.instance, event)
	if err != nil {
		return err
	}

	is.triggerUpdate()

	return nil
}

func newInstanceEvent(eventType core.InstanceEventType, origin core.Origin, instanceId string, status core.MachineStatus, payload core.InstanceEventPayload) core.InstanceEvent {
	return core.InstanceEvent{
		Id:         eventId(),
		Type:       eventType,
		Origin:     origin,
		Payload:    payload,
		InstanceId: instanceId,
		Status:     status,
		Reported:   false,
		Timestamp:  time.Now(),
	}
}

func (s *instanceState) PushInstancePrepareEvent(ctx context.Context, retries int) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstancePrepare,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusPreparing,
		core.InstanceEventPayload{
			Prepare: &core.InstancePrepareEventPayload{
				Retries: retries,
			},
		},
	)

	err = s.persistChange(event)
	if err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstancePreparedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstancePrepared,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			Prepared: &core.InstancePreparedEventPayload{},
		},
	)

	err = s.persistChange(event)
	if err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstancePreparationFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstancePreparationFailed,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			PreparationFailed: &core.InstancePreparationFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	s.instance.State.DesiredStatus = core.MachineStatusDestroying

	err = s.persistChange(event)

	return nil
}

func (s *instanceState) PushInstanceStartEvent(ctx context.Context, isRestart bool) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	var origin core.Origin
	if isRestart {
		origin = core.OriginRavel
	} else {
		origin = core.OriginUser
	}

	event := newInstanceEvent(
		core.InstanceStart,
		origin,
		s.instance.Id,
		core.MachineStatusStarting,
		core.InstanceEventPayload{
			Start: &core.InstanceStartEventPayload{
				IsRestart: isRestart,
			},
		},
	)

	s.instance.State.DesiredStatus = core.MachineStatusRunning

	if isRestart {
		s.instance.State.Restarts++
	} else {
		s.instance.State.Restarts = 0
	}

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStartFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstanceStartFailed,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			StartFailed: &core.InstanceStartFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStartedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstanceStarted,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusRunning,
		core.InstanceEventPayload{
			Started: &core.InstanceStartedEventPayload{},
		},
	)

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStopEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstanceStop,
		core.OriginUser,
		s.instance.Id,
		core.MachineStatusStopping,
		core.InstanceEventPayload{
			Stop: &core.InstanceStopEventPayload{},
		},
	)

	s.lastEvent = event
	s.instance.State.DesiredStatus = core.MachineStatusStopped
	s.instance.State.Restarts = 0

	if err = s.persistChange(event); err != nil {
		return
	}
	return nil
}

func (s *instanceState) PushInstanceExitedEvent(ctx context.Context, payload core.InstanceExitedEventPayload) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := newInstanceEvent(
		core.InstanceExited,
		core.OriginRavel,
		s.instance.Id,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			Exited: &payload,
		},
	)

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceDestroyEvent(ctx context.Context, origin core.Origin, reason string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceEvent{
		Id:     eventId(),
		Type:   core.InstanceDestroy,
		Origin: core.OriginRavel,
		Payload: core.InstanceEventPayload{
			Destroy: &core.InstanceDestroyEventPayload{
				Reason: reason,
			},
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusDestroying,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	s.instance.State.DesiredStatus = core.MachineStatusDestroyed

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceDestroyedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceEvent{
		Id:     eventId(),
		Type:   core.InstanceDestroyed,
		Origin: core.OriginRavel,
		Payload: core.InstanceEventPayload{
			Destroyed: &core.InstanceDestroyedEventPayload{},
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusDestroyed,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	if err = s.persistChange(event); err != nil {
		return
	}

	return nil
}
