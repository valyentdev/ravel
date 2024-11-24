package machinerunner

import (
	"sync"

	"github.com/valyentdev/ravel/core/errdefs"
)

type machineStore struct {
	mutex    sync.RWMutex
	machines map[string]*MachineRunner
}

type Store interface {
	GetMachine(id string) (*MachineRunner, error)
	Foreach(func(*MachineRunner))
	AddMachine(machine *MachineRunner)
	RemoveMachine(id string)
}

func NewStore() Store {
	return &machineStore{
		machines: make(map[string]*MachineRunner),
	}
}

func (s *machineStore) GetMachine(id string) (*MachineRunner, error) {
	s.mutex.RLock()
	machine, ok := s.machines[id]
	s.mutex.RUnlock()
	if !ok {
		return nil, errdefs.NewNotFound("machine not found")
	}

	return machine, nil
}

func (s *machineStore) AddMachine(machine *MachineRunner) {
	s.mutex.Lock()
	s.machines[machine.Id()] = machine
	s.mutex.Unlock()
}

func (s *machineStore) RemoveMachine(id string) {
	s.mutex.Lock()
	delete(s.machines, id)
	s.mutex.Unlock()
}

func (s *machineStore) Foreach(f func(*MachineRunner)) {
	s.mutex.RLock()
	for _, machine := range s.machines {
		f(machine)
	}
	s.mutex.RUnlock()
}
