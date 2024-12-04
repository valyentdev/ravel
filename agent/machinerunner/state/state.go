package state

import (
	"crypto/rand"
	"log/slog"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
)

type Eventer interface {
	ReportEvent(event api.MachineEvent)
}

func eventId() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}

type Store interface {
	CreateMachineInstance(mi structs.MachineInstance) error
	LoadMachineInstances() ([]structs.MachineInstance, error)
	DeleteMachineInstance(id string) error
	UpdateMachineInstanceState(id string, mi structs.MachineInstanceState) error
	DeleteMachineInstanceEvent(eventId string) error
	LoadMachineInstanceEvents() ([]api.MachineEvent, error)
	PutMachineInstanceEvent(event api.MachineEvent) error
}
type MachineInstanceState struct {
	id          string
	mutex       sync.RWMutex
	store       Store
	mi          structs.MachineInstance
	eventer     Eventer
	reportState func(mi cluster.MachineInstance) error
	stopCh      chan struct{}
	updateCh    chan struct{}
}

func (s *MachineInstanceState) triggerUpdate() {
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
	store Store,
	machine structs.MachineInstance,
	eventer Eventer,
	reportState func(mi cluster.MachineInstance) error,
) *MachineInstanceState {
	return newMachineInstanceState(store, machine, eventer, reportState)
}

func newMachineInstanceState(
	store Store,
	machine structs.MachineInstance,
	eventer Eventer,
	reportState func(mi cluster.MachineInstance) error,
) *MachineInstanceState {
	is := &MachineInstanceState{
		id:          machine.Machine.Id,
		store:       store,
		mi:          machine,
		eventer:     eventer,
		updateCh:    make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		reportState: reportState,
	}

	is.triggerUpdate()

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

func (is *MachineInstanceState) persistState() error {
	is.mi.State.UpdatedAt = time.Now()
	if err := is.store.UpdateMachineInstanceState(is.id, is.mi.State); err != nil {
		slog.Error("failed to update machine state in store", "machine_id", is.id, "error", err)
		return err
	}

	is.triggerUpdate()
	return nil
}

func (i *MachineInstanceState) UpdateStatus(status api.MachineStatus) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.mi.State.Status = status

	return i.persistState()
}

func (i *MachineInstanceState) UpdateDesiredStatus(status api.MachineStatus) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.mi.State.DesiredStatus = status

	return i.persistState()
}
