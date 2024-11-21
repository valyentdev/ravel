package machine

import (
	"sync"

	"github.com/valyentdev/ravel/core/errdefs"
)

type machineStore struct {
	mutex    sync.RWMutex
	machines map[string]*Machine
}

type Store interface {
	GetMachine(id string) (*Machine, error)
	Foreach(func(*Machine))
	AddMachine(machine *Machine)
	RemoveMachine(id string)
}

func NewStore() Store {
	return &machineStore{
		machines: make(map[string]*Machine),
	}
}

func (s *machineStore) GetMachine(id string) (*Machine, error) {
	s.mutex.RLock()
	machine, ok := s.machines[id]
	s.mutex.RUnlock()
	if !ok {
		return nil, errdefs.NewNotFound("machine not found")
	}

	return machine, nil
}

func (s *machineStore) AddMachine(machine *Machine) {
	s.mutex.Lock()
	s.machines[machine.Id()] = machine
	s.mutex.Unlock()
}

func (s *machineStore) RemoveMachine(id string) {
	s.mutex.Lock()
	delete(s.machines, id)
	s.mutex.Unlock()
}

func (s *machineStore) Foreach(f func(*Machine)) {
	s.mutex.RLock()
	for _, machine := range s.machines {
		f(machine)
	}
	s.mutex.RUnlock()
}
