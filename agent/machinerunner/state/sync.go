package state

import (
	"time"

	"github.com/valyentdev/ravel/api"
)

func (s *MachineInstanceState) startSyncing() {
	for range s.updateCh {
		finished, err := s.sync()
		if err != nil {
			go func() {
				time.Sleep(5 * time.Second)
				s.triggerUpdate()
			}()
			continue
		}

		if finished {
			return
		}
	}
}

func (s *MachineInstanceState) sync() (finished bool, err error) {
	s.mutex.RLock()
	cmi := s.mi.ClusterInstance()
	s.mutex.RUnlock()

	err = s.reportState(cmi)
	if err != nil {
		return false, err
	}

	if cmi.Status == api.MachineStatusDestroyed {
		return true, nil
	}

	return false, nil
}
