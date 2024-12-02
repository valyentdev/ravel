package state

import (
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
)

func (s *MachineInstanceState) startSyncing() {
	for range s.updateCh {
		slog.Debug("triggering machine instance state sync")
		finished, err := s.sync()
		if err != nil {
			slog.Debug("error syncing machine instance state", "err", err)
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
