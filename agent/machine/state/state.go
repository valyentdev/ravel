package state

import (
	"crypto/rand"
	"sync"
	"sync/atomic"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/agent/store"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
)

type StateReporter interface {
	UpsertInstanceState(instance cluster.MachineInstance) error
	DeleteInstanceState(id string) error
}

type Eventer interface {
	ReportEvent(event api.MachineEvent)
}

func eventId() ulid.ULID {
	return ulid.MustNew(ulid.Now(), rand.Reader)
}

type MachineInstanceState struct {
	id       string
	mutex    sync.RWMutex
	store    *store.Store
	mi       structs.MachineInstance
	eventer  Eventer
	sr       StateReporter
	stopCh   chan struct{}
	updateCh chan struct{}
	reported atomic.Bool
}

func (s *MachineInstanceState) triggerUpdate() {
	s.reported.Store(false)
	select {
	case s.updateCh <- struct{}{}:
		return
	default:
		<-s.updateCh
		s.updateCh <- struct{}{}
		return
	}
}

func NewMachineInstanceState(
	store *store.Store,
	machine structs.MachineInstance,
	eventer Eventer,
	sr StateReporter,
) *MachineInstanceState {
	return newMachineInstanceState(store, machine, eventer, false, sr)
}

func newMachineInstanceState(
	store *store.Store,
	machine structs.MachineInstance,
	eventer Eventer,
	new bool,
	stateReporter StateReporter,
) *MachineInstanceState {
	is := &MachineInstanceState{
		id:       machine.Machine.Id,
		store:    store,
		mi:       machine,
		eventer:  eventer,
		updateCh: make(chan struct{}, 1),
		stopCh:   make(chan struct{}),
		sr:       stateReporter,
	}

	if !new {
		go is.triggerUpdate()
	}

	go is.startSyncing()

	return is
}

func (i *MachineInstanceState) Id() string {
	return i.id
}

func (i *MachineInstanceState) InstanceId() string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.mi.Machine.InstanceId
}

func (i *MachineInstanceState) MachineInstance() structs.MachineInstance {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.mi
}

func (i *MachineInstanceState) State() structs.MachineInstanceState {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.mi.State
}

func (i *MachineInstanceState) Status() api.MachineStatus {
	i.mutex.RLock()
	defer i.mutex.RUnlock()
	return i.mi.State.Status
}
