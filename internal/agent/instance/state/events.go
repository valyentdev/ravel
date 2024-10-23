package state

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
)

func (is *instanceState) persistAndReportChange(event core.InstanceEvent) error {
	is.eventer.Report(event)

	is.lastEvent = event
	is.instance.State.Status = event.Status

	err := is.store.UpdateInstance(is.instance, event)
	if err != nil {
		return err
	}

	is.triggerUpdate()

	return nil
}

func (s *instanceState) newInstanceEvent(eventType core.InstanceEventType, origin core.Origin, status core.MachineStatus, payload core.InstanceEventPayload) core.InstanceEvent {
	return core.InstanceEvent{
		Id:         eventId(),
		Type:       eventType,
		Origin:     origin,
		Payload:    payload,
		InstanceId: s.instance.Id,
		MachineId:  s.instance.MachineId,
		Status:     status,
		Timestamp:  time.Now(),
	}
}

func (s *instanceState) PushInstancePrepareEvent(ctx context.Context, retries int) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstancePrepare,
		core.OriginRavel,
		core.MachineStatusPreparing,
		core.InstanceEventPayload{
			Prepare: &core.InstancePrepareEventPayload{
				Retries: retries,
			},
		},
	)

	err = s.persistAndReportChange(event)
	if err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstancePreparedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstancePrepared,
		core.OriginRavel,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			Prepared: &core.InstancePreparedEventPayload{},
		},
	)

	err = s.persistAndReportChange(event)
	if err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstancePreparationFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstancePreparationFailed,
		core.OriginRavel,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			PreparationFailed: &core.InstancePreparationFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	s.instance.State.DesiredStatus = core.MachineStatusDestroying

	err = s.persistAndReportChange(event)

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

	event := s.newInstanceEvent(
		core.InstanceStart,
		origin,
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

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStartFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceStartFailed,
		core.OriginRavel,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			StartFailed: &core.InstanceStartFailedEventPayload{
				Error: errMsg,
			},
		},
	)

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStartedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceStarted,
		core.OriginRavel,
		core.MachineStatusRunning,
		core.InstanceEventPayload{
			Started: &core.InstanceStartedEventPayload{},
		},
	)

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceStopEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceStop,
		core.OriginUser,
		core.MachineStatusStopping,
		core.InstanceEventPayload{
			Stop: &core.InstanceStopEventPayload{},
		},
	)

	s.lastEvent = event
	s.instance.State.DesiredStatus = core.MachineStatusStopped
	s.instance.State.Restarts = 0

	if err = s.persistAndReportChange(event); err != nil {
		return
	}
	return nil
}

func (s *instanceState) PushInstanceExitedEvent(ctx context.Context, payload core.InstanceExitedEventPayload) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceExited,
		core.OriginRavel,
		core.MachineStatusStopped,
		core.InstanceEventPayload{
			Exited: &payload,
		},
	)

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceDestroyEvent(ctx context.Context, origin core.Origin, force bool, reason string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceDestroy,
		origin,
		core.MachineStatusDestroying,
		core.InstanceEventPayload{
			Destroy: &core.InstanceDestroyEventPayload{
				Reason: reason,
				Force:  force,
			},
		},
	)

	s.instance.State.DesiredStatus = core.MachineStatusDestroyed

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	return nil
}

func (s *instanceState) PushInstanceDestroyedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceDestroyed,
		core.OriginRavel,
		core.MachineStatusDestroyed,
		core.InstanceEventPayload{
			Destroyed: &core.InstanceDestroyedEventPayload{},
		},
	)

	if err = s.persistAndReportChange(event); err != nil {
		return
	}

	s.eventer.Stop()

	return nil
}
