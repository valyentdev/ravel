package state

import (
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
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
	state := s.fsm.State()
	err = s.reportState(cluster.MachineInstance{
		Id:                   s.machine.InstanceId,
		Node:                 s.machine.Node,
		Namespace:            s.machine.Namespace,
		MachineVersion:       s.machine.MachineVersion,
		Status:               state.Status,
		Events:               state.LastEvents,
		LocalIPV4:            s.networking.Local.InstanceIP.String(),
		CreatedAt:            state.CreatedAt,
		UpdatedAt:            state.UpdatedAt,
		EnableMachineGateway: state.MachineGatewayEnabled,
	})
	if err != nil {
		return false, err
	}

	if state.Status == api.MachineStatusDestroyed {
		return true, nil
	}

	return false, nil
}
