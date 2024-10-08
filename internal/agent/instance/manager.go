package instance

import (
	"sync"

	"github.com/valyentdev/ravel/internal/agent/instance/state"
	"github.com/valyentdev/ravel/internal/agent/logging"
	"github.com/valyentdev/ravel/internal/agent/structs"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

type Manager struct {
	logger      *logging.InstanceLogger
	mutex       sync.Mutex
	state       state.InstanceState
	reservation structs.Reservation
	isPrepared  bool
	isRunning   bool
	waitCh      chan struct{}
	runtime     runtimes.Runtime
}

func NewInstanceManager(state state.InstanceState, runtime runtimes.Runtime, reservation structs.Reservation) *Manager {
	return &Manager{
		state:       state,
		runtime:     runtime,
		reservation: reservation,
		logger:      logging.NewInstanceLogger(state.Id()),
	}
}

func (m *Manager) Instance() core.Instance {
	return m.state.Instance()
}

func (m *Manager) Status() core.InstanceStatus {
	return m.state.Status()
}
