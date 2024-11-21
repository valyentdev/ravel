package instancemanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/logging"
)

type Manager struct {
	logger *logging.InstanceLogger
	state  *State

	vmBuilder instance.Builder
	vm        instance.VM
	er        instance.EventReporter

	vmLock     sync.RWMutex
	networking instance.NetworkingService

	waitCh chan struct{}
}

func (m *Manager) getVM() instance.VM {
	m.vmLock.RLock()
	vmm := m.vm
	m.vmLock.RUnlock()
	return vmm
}

func (m *Manager) setVM(vm instance.VM) {
	m.vmLock.Lock()
	m.vm = vm
	m.vmLock.Unlock()
}

func NewInstanceManager(
	store instance.InstanceStore,
	instance instance.Instance,
	ns instance.NetworkingService,
	vmBuilder instance.Builder,
	er instance.EventReporter,
) *Manager {
	state := NewState(store, instance)
	return &Manager{
		state:      state,
		logger:     logging.NewInstanceLogger(instance.Id),
		networking: ns,
		vmBuilder:  vmBuilder,
		er:         er,
	}
}

func (m *Manager) Instance() instance.Instance {
	return m.state.Instance()
}

func (m *Manager) Status() instance.InstanceStatus {
	return m.state.Status()
}

const defaultExecTimeout = time.Duration(5) * time.Second

func (m *Manager) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	if len(cmd) == 0 {
		return nil, errdefs.NewInvalidArgument("cmd is required")
	}

	if timeout == 0 {
		timeout = defaultExecTimeout
	}

	if timeout > 30*time.Second {
		return nil, errdefs.NewInvalidArgument("timeout must be less than 30 seconds")
	}

	vm := m.getVM()
	if vm == nil {
		return nil, errdefs.NewFailedPrecondition("instance is not running")
	}

	return vm.Exec(ctx, cmd, timeout)
}

func (m *Manager) GetLog() []*api.LogEntry {
	return m.logger.GetLog()
}

func (m *Manager) SubscribeToLogs() ([]*api.LogEntry, *logging.LogSubscriber) {
	return m.logger.Subscribe()
}

func (m *Manager) Signal(ctx context.Context, signal string) error {
	vm := m.getVM()
	if m.state.Status() != instance.InstanceStatusRunning || vm == nil {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is %s", m.Status()))
	}

	return vm.Signal(ctx, signal)
}
