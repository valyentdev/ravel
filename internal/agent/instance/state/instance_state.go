package state

import (
	"context"
	"crypto/rand"
	"sync"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/internal/agent/instance/eventer"
	"github.com/valyentdev/ravel/internal/agent/store"
	"github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/pkg/core"
)

func eventId() ulid.ULID {
	return ulid.MustNew(ulid.Now(), rand.Reader)
}

type instanceState struct {
	updateLock sync.Mutex
	cluster    *cluster.ClusterState
	store      *store.Store
	instance   core.Instance
	lastEvent  core.InstanceEvent
	updateCh   chan struct{}
	eventer    *eventer.Eventer
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
	return i.lastEvent
}

func NewInstanceState(
	state *store.Store,
	instance core.Instance,
	lastEvent core.InstanceEvent,
	nodeId string,
	c *cluster.ClusterState,
	eventer *eventer.Eventer,
) InstanceState {
	return newInstanceState(state, instance, lastEvent, c, eventer, false)
}

func newInstanceState(
	state *store.Store,
	instance core.Instance,
	lastEvent core.InstanceEvent,
	c *cluster.ClusterState,
	eventer *eventer.Eventer,
	new bool,
) *instanceState {
	is := &instanceState{
		store:    state,
		instance: instance,
		cluster:  c,
		updateCh: make(chan struct{}, 1),
		eventer:  eventer,
	}

	go is.sync()

	if !new {
		is.lastEvent = lastEvent
		go is.triggerUpdate()
	}

	return is
}

func CreateInstance(
	state *store.Store,
	cs *cluster.ClusterState,
	instance core.Instance,
	eventer *eventer.Eventer,
) (InstanceState, error) {
	is := newInstanceState(state, instance, core.InstanceEvent{}, cs, eventer, true)

	if err := is.create(); err != nil {
		return nil, err
	}

	return is, nil
}

func (s *instanceState) create() (err error) {
	s.updateLock.Lock()
	defer s.updateLock.Unlock()

	event := s.newInstanceEvent(
		core.InstanceCreated,
		core.OriginRavel,
		core.MachineStatusCreated,
		core.InstanceEventPayload{
			Created: &core.InstanceCreatedEventPayload{},
		},
	)

	err = s.store.CreateInstance(s.instance, event)
	if err != nil {
		return
	}
	s.lastEvent = event
	s.triggerUpdate()
	s.eventer.Report(event)
	return nil
}
