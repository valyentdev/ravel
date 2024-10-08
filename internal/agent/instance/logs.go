package instance

import (
	"github.com/valyentdev/ravel/internal/agent/logging"
	"github.com/valyentdev/ravel/pkg/core"
)

func (m *Manager) GetLog() []*core.LogEntry {
	return m.logger.GetLog()
}

func (m *Manager) SubscribeToLogs() logging.LogSubscriber {
	return m.logger.Subscribe()
}
