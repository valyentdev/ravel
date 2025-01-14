package runtime

import (
	"sync"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/instancerunner"
)

type State struct {
	mutex     sync.RWMutex
	instances map[string]*instancerunner.InstanceRunner

	idsMutex sync.RWMutex
	ids      map[string]struct{}
}

func NewState() *State {
	return &State{
		instances: make(map[string]*instancerunner.InstanceRunner),
		ids:       make(map[string]struct{}),
	}
}

func (s *State) List() []instance.Instance {
	s.mutex.RLock()
	instances := []instance.Instance{}
	for _, manager := range s.instances {
		instances = append(instances, manager.Instance())
	}
	s.mutex.RUnlock()
	return instances
}

func (s *State) AddInstance(id string, manager *instancerunner.InstanceRunner) {
	s.mutex.Lock()
	s.instances[id] = manager
	s.mutex.Unlock()
}

func (s *State) Delete(id string) {
	s.mutex.Lock()
	delete(s.instances, id)
	s.mutex.Unlock()
}

func (s *State) GetInstance(id string) (*instancerunner.InstanceRunner, error) {
	s.mutex.RLock()
	instance, ok := s.instances[id]
	s.mutex.RUnlock()
	if !ok {
		return nil, errdefs.NewNotFound("instance not found")
	}
	return instance, nil
}

func (s *State) ReserveId(id string) bool {
	s.idsMutex.Lock()
	if _, ok := s.ids[id]; ok {
		return false
	}
	s.ids[id] = struct{}{}
	s.idsMutex.Unlock()
	return true
}

func (s *State) ReleaseId(id string) {
	s.idsMutex.Lock()
	delete(s.ids, id)
	s.idsMutex.Unlock()
}
