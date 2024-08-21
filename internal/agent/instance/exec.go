package instance

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

const defaultTimeout = 10 * time.Second

func (m *Manager) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*runtimes.ExecResult, error) {
	if !m.isRunning {
		return nil, core.NewFailedPrecondition("instance is not running")
	}

	var realTimeout time.Duration
	if timeout == 0 {
		realTimeout = defaultTimeout
	} else {
		realTimeout = timeout
	}

	return m.runtime.Exec(ctx, m.state.Instance().Id, cmd, realTimeout)
}
