package machinerunner

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
)

func (m *MachineRunner) Run() {
	status := m.state.State().Status
	if status == api.MachineStatusStopping || status == api.MachineStatusRunning {
		go m.runInstance()
	}
	if status == api.MachineStatusCreated {
		m.state.PushPrepareEvent()
	}

	for event := range m.state.Events() {
		slog.Debug("machine event", "machine", event.MachineId, "event", event.Type, "new_status", event.Status)
		eventType := event.Type
		switch eventType {
		case api.MachinePrepare:
			go m.prepare(context.Background())
		case api.MachinePrepared:
			if m.state.State().DesiredStatus == api.MachineStatusRunning {
				m.state.PushStartEvent(false)
			}
		case api.MachinePrepareFailed:
			m.state.PushDestroyEvent(api.OriginRavel, true, false, "failed to prepare machine")
		case api.MachineExited:
			m.handleExit(event.Payload.Exited)
		case api.MachineStarted:
			go m.runInstance()
		case api.MachineStart:
			go m.startInstance()
		case api.MachineStop:
			go m.stopInstance(context.Background(), event.Payload.Stop.Config)
		case api.MachineDestroy:
			go m.destroyImpl()
		case api.MachineDestroyed:
			m.onDestroyed(m.state.MachineInstance())
			return
		}
	}
}

func (m *MachineRunner) runInstance() error {
	ctx, cancel := context.WithCancel(context.Background())
	updates, err := m.runtime.WatchInstanceState(ctx, m.state.InstanceId())
	if err != nil {
		cancel()
		return err
	}
	go func() {
		defer cancel()
		for update := range updates {
			if update.Status == instance.InstanceStatusStopped {
				var payload api.MachineExitedEventPayload
				if update.ExitResult != nil {
					payload.ExitCode = update.ExitResult.ExitCode
					payload.ExitedAt = update.ExitResult.ExitedAt
				} else {
					payload.ExitCode = -1
					payload.ExitedAt = time.Now()
				}

				m.state.PushExitedEvent(payload)
			}
		}
	}()

	return nil
}

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

	restarts := state.Restarts
	restartIn := 1 * time.Second
	if restarts > 1 {
		restartIn = 5 * time.Second
	}
	go m.autoRestartIn(restartIn)
}

func (m *MachineRunner) autoRestartIn(duration time.Duration) {
	time.Sleep(duration)
	if m.state.State().DesiredStatus != api.MachineStatusRunning {
		return
	}
	m.state.PushStartEvent(true)
}

func (m *MachineRunner) handleExitWithAutoDestroy(p *api.MachineExitedEventPayload) {
	state := m.state.State()
	success := p.ExitCode == 0
	if success {
		m.state.PushDestroyEvent(api.OriginRavel, false, true, "machine auto-destroyed after successful exit")
		return
	}
	restartConfig := m.state.MachineInstance().Version.Config.Workload.Restart

	if restartConfig.Policy == api.RestartPolicyNever {
		m.state.PushDestroyEvent(api.OriginRavel, false, true, "machine auto-destroyed after failed exit")
		return
	}

	if restartConfig.Policy == api.RestartPolicyOnFailure {
		if state.Restarts >= restartConfig.MaxRetries {
			m.state.PushDestroyEvent(api.OriginRavel, false, true, "machine auto-destroyed after failed exit")
			return
		}
	}

	restarts := state.Restarts
	restartIn := 5 * time.Second * time.Duration(restarts)
	go m.autoRestartIn(restartIn)
}
