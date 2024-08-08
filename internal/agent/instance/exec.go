package instance

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/internal/agent/runtimes"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

const defaultTimeout = 10 * time.Second

func (m *Manager) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*runtimes.ExecResult, error) {
	if !m.isRunning {
		return nil, ravelerrors.ErrInstanceIsNotRunning
	}

	var realTimeout time.Duration
	if timeout == 0 {
		realTimeout = defaultTimeout
	} else {
		realTimeout = timeout
	}

	return m.runtime.Exec(ctx, m.state.Instance().Id, cmd, realTimeout)
}
