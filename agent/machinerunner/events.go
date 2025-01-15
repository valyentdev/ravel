package machinerunner

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/api"
)

func (m *MachineRunner) handleExit(p *api.MachineExitedEventPayload) {
	state := m.state.State()

	config := m.state.MachineInstance().Version.Config
	if config.Workload.AutoDestroy {
		m.handleExitWithAutoDestroy(p)
		return
	}

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

func (m *MachineRunner) handleExitWithAutoDestroy(p *api.MachineExitedEventPayload) {
	success := p.ExitCode == 0
	if success {
		err := m.destroyImpl(context.Background(), destroyPayload{
			origin:      api.OriginRavel,
			reason:      "machine destroyed after successful exit",
			autoDestroy: true,
		})
		slog.Error("failed to destroy machine", "err", err)
		return
	}
	restartConfig := m.state.MachineInstance().Version.Config.Workload.Restart

	if restartConfig.Policy == api.RestartPolicyNever {
		err := m.destroyImpl(context.Background(), destroyPayload{
			origin:      api.OriginRavel,
			reason:      "machine destroyed after failed exit",
			autoDestroy: true,
		})
		slog.Error("failed to destroy machine", "err", err)
		return
	}

	if restartConfig.Policy == api.RestartPolicyOnFailure {
		if m.state.State().Restarts >= restartConfig.MaxRetries {
			err := m.destroyImpl(context.Background(), destroyPayload{
				origin:      api.OriginRavel,
				reason:      "machine destroyed after failed exit and max retries",
				autoDestroy: true,
			})
			slog.Error("failed to destroy machine", "err", err)
			return
		}
	}

	m.start(true)

}
