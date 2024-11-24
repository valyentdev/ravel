package machinerunner

import (
	"github.com/valyentdev/ravel/api"
)

func (m *MachineRunner) handleExit(p *api.MachineExitedEventPayload) {
	state := m.state.State()
	if state.DesiredStatus != api.MachineStatusRunning {
		return
	}

	restartConfig := m.state.MachineInstance().Version.Config.Workload.Restart

	if restartConfig.Policy == api.RestartPolicyNever {
		m.state.UpdateDesiredStatus(api.MachineStatusStopped)
		return
	}

	if restartConfig.Policy == api.RestartPolicyOnFailure {
		if state.Restarts >= restartConfig.MaxRetries || p.ExitCode == 0 {
			m.state.UpdateDesiredStatus(api.MachineStatusStopped)
			return
		}
	}

	m.start(true)
}
