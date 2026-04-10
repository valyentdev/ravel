package machinerunner

import (
	"context"
	"sync"

	"github.com/alexisbouchez/ravel/agent/machinerunner/state"
	"github.com/alexisbouchez/ravel/agent/structs"
	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/core/daemon"
)

type MachineRunner struct {
	state       *state.MachineInstanceState
	runtime     daemon.Runtime
	runLock     sync.Mutex
	onDestroyed func(m structs.MachineInstance)
}

func (m *MachineRunner) Id() string {
	return m.state.Id()
}

func New(
	store state.Store,
	machine structs.MachineInstance,
	runtime daemon.Runtime,
	reportState func(mi cluster.MachineInstance) error,
	eventer state.Eventer,
	onDestroyed func(m structs.MachineInstance),
) *MachineRunner {
	m := &MachineRunner{
		state:       state.NewMachineInstanceState(store, machine, eventer, reportState),
		runtime:     runtime,
		onDestroyed: onDestroyed,
	}

	return m
}

func (m *MachineRunner) WaitForStatus(ctx context.Context, status api.MachineStatus) error {
	return m.state.WaitForStatus(ctx, status)
}
