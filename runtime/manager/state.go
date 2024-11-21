package instancemanager

import (
	"log/slog"
	"sync"

	"github.com/valyentdev/ravel/core/instance"
)

type State struct {
	store instance.InstanceStore
	mutex sync.Mutex

	instance     instance.Instance
	instanceLock sync.RWMutex
}

func NewState(store instance.InstanceStore, instance instance.Instance) *State {
	return &State{
		store:    store,
		instance: instance,
	}
}

func (s *State) Instance() instance.Instance {
	s.instanceLock.RLock()
	defer s.instanceLock.RUnlock()
	return s.instance
}

func (s *State) Status() instance.InstanceStatus {
	s.instanceLock.RLock()
	defer s.instanceLock.RUnlock()
	return s.instance.State.Status
}

func (s *State) Lock() {
	s.mutex.Lock()
}

func (s *State) Unlock() {
	s.mutex.Unlock()
}

func (s *State) UpdateInstanceState(state instance.State) error {
	s.instanceLock.Lock()
	defer s.instanceLock.Unlock()
	s.instance.State = state
	err := s.store.PutInstance(s.instance)
	if err != nil {
		slog.Error("failed to update instance state", "error", err)
		return err
	}
	return nil
}
