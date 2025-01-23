package state

import (
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/internal/sm"
)

const maxSize = 10

type MachineState = structs.MachineInstanceState

func (s *MachineInstanceState) newEvent(eventType api.MachineEventType, origin api.Origin, status api.MachineStatus, payload api.MachineEventPayload) *api.MachineEvent {
	return &api.MachineEvent{
		Id:         eventId(),
		Type:       eventType,
		Origin:     origin,
		Payload:    payload,
		MachineId:  s.machine.Id,
		InstanceId: s.machine.InstanceId,
		Status:     status,
		Timestamp:  time.Now(),
	}
}

func (s *MachineInstanceState) PushPrepareEvent() (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachinePrepare,
		api.OriginRavel,
		api.MachineStatusPreparing,
		api.MachineEventPayload{},
	)

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushPreparedEvent() (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachinePrepared,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{},
	)
	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushPrepareFailedEvent(errMsg string) (prev, next *MachineState, _ error) {
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

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushStartEvent(isRestart bool) (prev, next *MachineState, _ error) {
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

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushStartFailedEvent(errMsg string) (prev, next *MachineState, _ error) {
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

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushStartedEvent() (prev, next *MachineState, _ error) {
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

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushStopEvent(payload api.MachineStopEventPayload) (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachineStop,
		api.OriginUser,
		api.MachineStatusStopping,
		api.MachineEventPayload{
			Stop: &payload,
		},
	)

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushStopFailedEvent() (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachineStopFailed,
		api.OriginRavel,
		"",
		api.MachineEventPayload{},
	)

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushExitedEvent(payload api.MachineExitedEventPayload) (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachineExited,
		api.OriginRavel,
		api.MachineStatusStopped,
		api.MachineEventPayload{
			Exited: &payload,
		},
	)

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushDestroyEvent(origin api.Origin, force bool, autoDestroy bool, reason string) (prev, next *MachineState, _ error) {
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

	return s.pushEvent(event)
}

func (s *MachineInstanceState) PushDestroyedEvent() (prev, next *MachineState, _ error) {
	event := s.newEvent(
		api.MachineDestroyed,
		api.OriginRavel,
		api.MachineStatusDestroyed,
		api.MachineEventPayload{},
	)

	return s.pushEvent(event)
}

func (s *MachineInstanceState) EnableGateway() error {
	return s.fsm.Mutate(func(mis *structs.MachineInstanceState) {
		mis.MachineGatewayEnabled = true
	})
}

func (s *MachineInstanceState) DisableGateway() error {
	return s.fsm.Mutate(func(mis *structs.MachineInstanceState) {
		mis.MachineGatewayEnabled = false
	})
}

func copyStatefunc(mis *structs.MachineInstanceState) *structs.MachineInstanceState {
	return &structs.MachineInstanceState{
		DesiredStatus:         mis.DesiredStatus,
		Status:                mis.Status,
		Restarts:              mis.Restarts,
		LastEvents:            mis.LastEvents,
		CreatedAt:             mis.CreatedAt,
		UpdatedAt:             mis.UpdatedAt,
		LocalIPV4:             mis.LocalIPV4,
		MachineGatewayEnabled: mis.MachineGatewayEnabled,
	}
}

type stateMachine = sm.StateMachine[structs.MachineInstanceState, api.MachineEventType, *api.MachineEvent]

func getLastEvents(currentEvents []api.MachineEvent, newEvent *api.MachineEvent) []api.MachineEvent {
	lastEvents := currentEvents
	currentSize := len(lastEvents)
	newSize := currentSize + 1
	if currentSize == maxSize {
		lastIndex := currentSize - 1
		lastEvents = lastEvents[:lastIndex-1]
		newSize = maxSize
	}

	newEvents := make([]api.MachineEvent, 0, newSize)
	newEvents = append(newEvents, *newEvent)
	newEvents = append(newEvents, lastEvents...)

	return newEvents
}

func (ms *MachineInstanceState) afterAll(mis *structs.MachineInstanceState, me *api.MachineEvent) error {
	mis.LastEvents = getLastEvents(mis.LastEvents, me)
	mis.Status = me.Status
	mis.UpdatedAt = me.Timestamp

	if err := ms.store.UpdateMachineInstance(ms.machine.Id, mis, me); err != nil {
		return err
	}

	slog.Info("machine state updated", "machine_id", ms.machine.Id, "status", mis.Status)

	ms.events <- me

	ms.triggerUpdate()

	return nil
}

func (ms *MachineInstanceState) afterMutate(mis *structs.MachineInstanceState) error {
	if err := ms.store.UpdateMachineInstance(ms.machine.Id, mis, nil); err != nil {
		return err
	}

	ms.triggerUpdate()

	return nil
}

func canStart(mis *structs.MachineInstanceState, _ *api.MachineEvent) bool {
	return mis.Status == api.MachineStatusStopped
}

func applyStart(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	if me.Payload.Start != nil && me.Payload.Start.IsRestart {
		mis.Restarts++
	} else {
		mis.Restarts = 0
	}
}

func canDestroy(mis *structs.MachineInstanceState, event *api.MachineEvent) bool {
	return (mis.Status == api.MachineStatusStopped || mis.Status == api.MachineStatusDestroying) || (mis.Status == api.MachineStatusRunning && event.Payload.Destroy.Force)
}

func applyDestroy(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	mis.DesiredStatus = api.MachineStatusDestroyed
}

func canPrepare(mis *structs.MachineInstanceState, _ *api.MachineEvent) bool {
	return mis.Status == api.MachineStatusCreated || mis.Status == api.MachineStatusPreparing
}

func applyPrepareFailed(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	mis.DesiredStatus = api.MachineStatusDestroyed
}

func canStop(mis *structs.MachineInstanceState, _ *api.MachineEvent) bool {
	return mis.Status == api.MachineStatusRunning || mis.Status == api.MachineStatusStopping
}

func applyStop(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	mis.DesiredStatus = api.MachineStatusStopped
	mis.Restarts = 0
}

func applyStopFailed(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	me.Status = mis.Status // Stop failed events don't change the status
}

func applyExited(mis *structs.MachineInstanceState, me *api.MachineEvent) {
	if mis.Status == api.MachineStatusDestroying || mis.Status == api.MachineStatusDestroyed {
		me.Status = mis.Status // When machine is force destroying/destroyed, we don't want to change the status
	}
}

func newFSM(initial *structs.MachineInstanceState, afterAll func(mis *structs.MachineInstanceState, me *api.MachineEvent) error, afterMutate func(mis *structs.MachineInstanceState) error) *stateMachine {
	sm := sm.NewStateMachine(
		initial,
		sm.Config[structs.MachineInstanceState, api.MachineEventType, *api.MachineEvent]{
			Copy: copyStatefunc,
			ApplyAll: func(mis *structs.MachineInstanceState, me *api.MachineEvent) {
				mis.UpdatedAt = me.Timestamp
				mis.Status = me.Status
			},
			AfterMutate:   afterMutate,
			AfterAllEvent: afterAll,
			Transitions: sm.Transitions[structs.MachineInstanceState, api.MachineEventType, *api.MachineEvent]{
				api.MachineCreated: {},
				api.MachinePrepare: {
					Can: canPrepare,
				},
				api.MachinePrepared: {},
				api.MachinePrepareFailed: {
					Apply: applyPrepareFailed,
				},
				api.MachineStart: {
					Can:   canStart,
					Apply: applyStart,
				},
				api.MachineStarted:     {},
				api.MachineStartFailed: {},
				api.MachineStop: {
					Can:   canStop,
					Apply: applyStop,
				},
				api.MachineStopFailed: {
					Apply: applyStopFailed,
				},
				api.MachineExited: {
					Apply: applyExited,
				},
				api.MachineDestroy: {
					Can:   canDestroy,
					Apply: applyDestroy,
				},
				api.MachineDestroyed: {},
			},
		},
	)

	return sm
}
