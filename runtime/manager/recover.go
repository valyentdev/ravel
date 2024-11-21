package instancemanager

import (
	"context"
	"log/slog"

	"github.com/valyentdev/ravel/core/instance"
)

func (m *Manager) Recover() {
	status := m.state.Status()
	if status == instance.InstanceStatusRunning || status == instance.InstanceStatusStarting || status == instance.InstanceStatusStopping {
		m.recoverRunning(status)
		return
	}
}

func (m *Manager) recoverRunning(status instance.InstanceStatus) {
	i := m.state.Instance()
	vmm, h, err := m.vmBuilder.RecoverInstanceVM(context.Background(), &i)
	if err != nil {
		err := m.state.UpdateInstanceState(instance.State{
			Status: instance.InstanceStatusStopped,
		})
		if err != nil {
			slog.Error("failed to update instance state", "error", err)
		}
		return
	}

	if status != instance.InstanceStatusRunning {
		err := m.state.UpdateInstanceState(instance.State{
			Status: instance.InstanceStatusStarting,
		})
		if err != nil {
			slog.Error("failed to update instance state", "error", err)
		}
	}

	m.waitCh = make(chan struct{})

	m.vmLock.Lock()
	m.vm = vmm
	m.vmLock.Unlock()

	go m.run(h)
}
