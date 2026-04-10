package machinerunner

import (
	"context"
	"log/slog"
	"time"

	"github.com/alexisbouchez/ravel/api"
)

const (
	defaultHealthCheckInterval = 30 // seconds
	defaultHealthCheckTimeout  = 5  // seconds
	defaultHealthCheckRetries  = 3
)

func (m *MachineRunner) startHealthCheck() {
	config := m.state.MachineInstance().Version.Config.Workload.HealthCheck
	if config == nil || len(config.Exec) == 0 {
		return
	}

	interval := defaultHealthCheckInterval
	if config.Interval > 0 {
		interval = config.Interval
	}

	timeout := defaultHealthCheckTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	retries := defaultHealthCheckRetries
	if config.Retries > 0 {
		retries = config.Retries
	}

	go m.runHealthCheck(config.Exec, interval, timeout, retries)
}

func (m *MachineRunner) runHealthCheck(cmd []string, interval, timeout, maxRetries int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	consecutiveFailures := 0
	currentHealth := api.HealthStatusStarting

	m.state.UpdateHealth(currentHealth)

	for {
		select {
		case <-ticker.C:
			status := m.state.State().Status
			if status != api.MachineStatusRunning {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			result, err := m.Exec(ctx, cmd, time.Duration(timeout)*time.Second)
			cancel()

			if err != nil || result.ExitCode != 0 {
				consecutiveFailures++
				slog.Debug("health check failed",
					"machine", m.state.Id(),
					"failures", consecutiveFailures,
					"err", err)

				if consecutiveFailures >= maxRetries {
					if currentHealth != api.HealthStatusUnhealthy {
						currentHealth = api.HealthStatusUnhealthy
						m.state.UpdateHealth(currentHealth)
						slog.Warn("machine unhealthy",
							"machine", m.state.Id(),
							"consecutive_failures", consecutiveFailures)
					}
				}
			} else {
				consecutiveFailures = 0
				if currentHealth != api.HealthStatusHealthy {
					currentHealth = api.HealthStatusHealthy
					m.state.UpdateHealth(currentHealth)
					slog.Info("machine healthy", "machine", m.state.Id())
				}
			}
		}
	}
}
