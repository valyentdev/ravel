package instance

import (
	"sync"

	"github.com/valyentdev/ravel/internal/agent/instance/state"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

type Manager struct {
	mutex      sync.Mutex
	state      state.InstanceState
	isPrepared bool
	isRunning  bool
	waitCh     chan struct{}
	runtime    runtimes.Runtime
}

func NewInstanceManager(state state.InstanceState, runtime runtimes.Runtime) *Manager {
	return &Manager{
		state:   state,
		runtime: runtime,
	}
}

func (m *Manager) Instance() core.Instance {
	return m.state.Instance()
}

func (m *Manager) Status() core.InstanceStatus {
	return m.state.Status()
}
