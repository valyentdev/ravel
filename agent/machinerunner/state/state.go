package state

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/internal/sm"
)

type Eventer interface {
	ReportEvent(event *api.MachineEvent)
}

func eventId() string {
	return ulid.MustNew(ulid.Now(), rand.Reader).String()
}

type Store interface {
	CreateMachineInstance(mi structs.MachineInstance) error
	LoadMachineInstances() ([]structs.MachineInstance, error)
	DeleteMachineInstance(id string) error
	UpdateMachineInstance(id string, mi *structs.MachineInstanceState, event *api.MachineEvent) error
	DeleteMachineInstanceEvent(eventId string) error
	LoadMachineInstanceEvents() ([]api.MachineEvent, error)
}
type MachineInstanceState struct {
	machine        cluster.Machine
	machineVersion api.MachineVersion
	networking     instance.NetworkingConfig
	store          Store
	eventer        Eventer
	fsm            *stateMachine
	reportState    func(mi cluster.MachineInstance) error
	updateCh       chan struct{}
	events         chan *api.MachineEvent
}

func (s *MachineInstanceState) pushEvent(event *api.MachineEvent) (prev, new *structs.MachineInstanceState, err error) {
	prev, new, err = s.fsm.PushEvent(event)
	if err == nil {
		s.eventer.ReportEvent(event)
		return prev, new, nil
	}
	if err == sm.ErrCannotTransition {
		return prev, new, errdefs.NewFailedPrecondition(fmt.Sprintf("machine is in %s status", prev.Status))
	}
	slog.Error("failed to push event", "error", err)
	return prev, new, err
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
	return newMachineInstanceState(store, &machine, eventer, reportState)
}

func newMachineInstanceState(
	store Store,
	machine *structs.MachineInstance,
	eventer Eventer,
	reportState func(mi cluster.MachineInstance) error,
) *MachineInstanceState {

	is := &MachineInstanceState{
		store:          store,
		eventer:        eventer,
		updateCh:       make(chan struct{}, 1),
		reportState:    reportState,
		machine:        machine.Machine,
		machineVersion: machine.Version,
		networking:     machine.Network,
		events:         make(chan *api.MachineEvent, 5),
	}

	if len(machine.State.LastEvents) > 0 {
		is.events <- &machine.State.LastEvents[0]
	}

	is.fsm = newFSM(&machine.State, is.afterAll, is.afterMutate)

	is.triggerUpdate()

	go is.startSyncing()

	return is
}

func (i *MachineInstanceState) Id() string {
	return i.machine.Id
}

func (i *MachineInstanceState) InstanceId() string {
	return i.machine.InstanceId
}

func (i *MachineInstanceState) MachineInstance() structs.MachineInstance {
	return structs.MachineInstance{
		Machine: i.machine,
		Version: i.machineVersion,
		Network: i.networking,
		State:   *i.State(),
	}
}

func (i *MachineInstanceState) State() *structs.MachineInstanceState {
	return i.fsm.State()
}

func (i *MachineInstanceState) Status() api.MachineStatus {
	return i.fsm.State().Status
}

func (i *MachineInstanceState) UpdateDesiredStatus(status api.MachineStatus) error {
	err := i.fsm.Mutate(func(mis *structs.MachineInstanceState) {
		mis.DesiredStatus = status
	})

	if err != nil {
		return err
	}

	return nil
}

func (i *MachineInstanceState) WaitForStatus(ctx context.Context, status api.MachineStatus) error {
	sub := i.fsm.Subscribe()
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case state := <-sub.Ch():
			if state.Status == status {
				return nil
			}
		}
	}
}

func (i *MachineInstanceState) Events() <-chan *api.MachineEvent {
	return i.events
}
