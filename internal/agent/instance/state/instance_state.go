package state

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/valyentdev/ravel/internal/agent/store"
	"github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/pkg/core"
)

func eventId() string {
	return id.GeneratePrefixed("event")
}

type instanceState struct {
	updateLock sync.Mutex
	cluster    *cluster.ClusterState
	store      *store.Store
	instance   core.Instance
	lastEvent  *core.InstanceEvent
	updateCh   chan struct{}
	node       string
}

func (s *instanceState) triggerUpdate() {
	select {
	case s.updateCh <- struct{}{}:
		return
	default:
		<-s.updateCh
		s.updateCh <- struct{}{}
		return
	}
}

type InstanceState interface {
	Id() string
	Instance() core.Instance
	Status() core.MachineStatus
	LastEvent() core.InstanceEvent
	Create(ctx context.Context) error
	PushInstancePrepareEvent(ctx context.Context, retries int) error
	PushInstancePreparedEvent(ctx context.Context) error
	PushInstancePreparationFailedEvent(ctx context.Context, errMsg string) error
	PushInstanceStartEvent(ctx context.Context, isRestart bool) error
	PushInstanceStartedEvent(ctx context.Context) error
	PushInstanceStartFailedEvent(ctx context.Context, errMsg string) error
	PushInstanceExitedEvent(ctx context.Context, payload core.InstanceExitedEventPayload) error
	PushInstanceStopEvent(ctx context.Context) error
	PushInstanceDestroyEvent(ctx context.Context, origin core.Origin, reason string) error
	PushInstanceDestroyedEvent(ctx context.Context) error
}

func (i *instanceState) Id() string {
	return i.instance.Id
}

func (i *instanceState) Instance() core.Instance {
	return i.instance
}

func (i *instanceState) Status() core.MachineStatus {
	return i.lastEvent.Status
}

func (i *instanceState) LastEvent() core.InstanceEvent {
	return *i.lastEvent
}

func NewInstanceState(state *store.Store, instance core.Instance, lastEvent *core.InstanceEvent, nodeId string, c *cluster.ClusterState) InstanceState {
	is := &instanceState{
		store:     state,
		instance:  instance,
		lastEvent: lastEvent,
		cluster:   c,
		node:      nodeId,
		updateCh:  make(chan struct{}, 1),
	}

	go is.sync()

	if lastEvent != nil {
		defer is.triggerUpdate()
	}

	return is
}

func (s *instanceState) Create(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	if s.lastEvent != nil && s.Status() != core.MachineStatusCreated {
		return nil
	}

	slog.Info("Creating instance", "instance", s.instance.Id)

	event := core.InstanceCreatedEvent{
		Id:         eventId(),
		Type:       core.InstanceCreated,
		Origin:     core.OriginRavel,
		Payload:    core.InstanceCreatedEventPayload{},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusCreated,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.CreateInstance(ctx, s.instance); err != nil {
		slog.Error("Failed to create instance", "instance", s.instance.Id, "error", err)
		return
	}

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		slog.Error("Failed to store instance event", "instance", s.instance.Id, "error", err)
		return
	}

	if err = tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction", "error", err)
		return
	}

	s.lastEvent = event.ToAny()

	slog.Info("Instance created", "instance", s.instance.Id, "event", s.lastEvent)

	return nil
}

func (s *instanceState) PushInstancePrepareEvent(ctx context.Context, retries int) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	if s.Status() == core.MachineStatusPreparing {
		return nil
	}

	event := core.InstancePrepareEvent{
		Id:     eventId(),
		Type:   core.InstancePrepare,
		Origin: core.OriginRavel,
		Payload: core.InstancePrepareEventPayload{
			Retries: retries,
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusPreparing,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	if err = s.store.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstancePreparedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstancePreparedEvent{
		Id:         eventId(),
		Type:       core.InstancePrepared,
		Origin:     core.OriginRavel,
		Payload:    core.InstancePreparedEventPayload{},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusStopped,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = s.store.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.MarkInstanceAsPrepared(ctx, s.instance.Id); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.instance.Prepared = true
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstancePreparationFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstancePreparationFailedEvent{
		Id:     eventId(),
		Type:   core.InstancePreparationFailed,
		Origin: core.OriginRavel,
		Payload: core.InstancePreparationFailedEventPayload{
			Error: errMsg,
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusCreated,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.UpdateInstanceDesiredStatus(ctx, s.instance.Id, core.MachineStatusDestroying); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceStartEvent(ctx context.Context, isRestart bool) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	status := s.Status()
	if status == core.MachineStatusStarting || status == core.MachineStatusRunning {
		return nil
	}

	var origin core.Origin
	if isRestart {
		origin = core.OriginRavel
	} else {
		origin = core.OriginUser
	}

	event := core.InstanceStartEvent{
		Id:     eventId(),
		Type:   core.InstanceStart,
		Origin: origin,
		Payload: core.InstanceStartEventPayload{
			IsRestart: isRestart,
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusStarting,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.UpdateInstanceDesiredStatus(ctx, s.instance.Id, core.MachineStatusRunning); err != nil {
		return
	}

	if isRestart {
		if err = tx.IncrementInstanceRestarts(ctx, s.instance.Id); err != nil {
			return
		}

	} else {
		if s.instance.Restarts > 0 {
			if err = tx.ResetRestarts(ctx, s.instance.Id); err != nil {
				return
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.instance.DesiredStatus = core.MachineStatusRunning
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceStartFailedEvent(ctx context.Context, errMsg string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceStartFailedEvent{
		Id:     eventId(),
		Type:   core.InstanceStartFailed,
		Origin: core.OriginRavel,
		Payload: core.InstanceStartFailedEventPayload{
			Error: errMsg,
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusStopped,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceStartedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceStartedEvent{
		Id:         eventId(),
		Type:       core.InstanceStarted,
		Origin:     core.OriginRavel,
		Payload:    core.InstanceStartedEventPayload{},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusRunning,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	if err = s.store.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceStopEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceStopEvent{
		Id:         eventId(),
		Type:       core.InstanceStop,
		Origin:     core.OriginRavel,
		Payload:    core.InstanceStopEventPayload{},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusStopping,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.UpdateInstanceDesiredStatus(ctx, s.instance.Id, core.MachineStatusStopped); err != nil {
		return
	}

	if err = tx.ResetRestarts(ctx, s.instance.Id); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.instance.DesiredStatus = core.MachineStatusStopped
	s.instance.Restarts = 0
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceExitedEvent(ctx context.Context, payload core.InstanceExitedEventPayload) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceExitedEvent{
		Id:         eventId(),
		Type:       core.InstanceExited,
		Origin:     core.OriginRavel,
		Payload:    payload,
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusStopped,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}
	s.lastEvent = event.ToAny()
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceDestroyEvent(ctx context.Context, origin core.Origin, reason string) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	status := s.Status()
	if status == core.MachineStatusDestroying || status == core.MachineStatusDestroyed {
		return nil
	}

	event := core.InstanceDestroyEvent{
		Id:     eventId(),
		Type:   core.InstanceDestroy,
		Origin: core.OriginRavel,
		Payload: core.InstanceDestroyEventPayload{
			Reason: reason,
		},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusDestroying,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.UpdateInstanceDesiredStatus(ctx, s.instance.Id, core.MachineStatusDestroyed); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.instance.DesiredStatus = core.MachineStatusDestroyed
	s.triggerUpdate()

	return nil
}

func (s *instanceState) PushInstanceDestroyedEvent(ctx context.Context) (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := core.InstanceDestroyedEvent{
		Id:         eventId(),
		Type:       core.InstanceDestroyed,
		Origin:     core.OriginRavel,
		Payload:    core.InstanceDestroyedEventPayload{},
		InstanceId: s.instance.Id,
		Status:     core.MachineStatusDestroyed,
		Reported:   false,
		Timestamp:  time.Now(),
	}

	tx, err := s.store.BeginTx(ctx)
	if err != nil {
		return
	}

	if err = tx.StoreInstanceEvent(ctx, event.ToAny()); err != nil {
		return
	}

	if err = tx.MarkInstanceDestroyed(ctx, s.instance.Id); err != nil {
		return
	}

	if err = tx.Commit(); err != nil {
		return
	}

	s.lastEvent = event.ToAny()
	s.instance.Destroyed = true
	s.triggerUpdate()

	return nil
}
