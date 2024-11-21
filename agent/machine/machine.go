package machine

import (
	"sync"
	"sync/atomic"

	"github.com/valyentdev/ravel/agent/machine/state"
	"github.com/valyentdev/ravel/agent/store"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/runtime"
)

type Machine struct {
	mutex   sync.Mutex
	state   *state.MachineInstanceState
	runtime *runtime.Runtime

	destroying atomic.Bool

	onDestroyed func(m structs.MachineInstance)
}

func (m *Machine) Id() string {
	return m.state.Id()
}

func NewMachine(
	store *store.Store,
	machine structs.MachineInstance,
	runtime *runtime.Runtime,
	sr state.StateReporter,
	eventer state.Eventer,
	onDestroyed func(m structs.MachineInstance),
) *Machine {
	m := &Machine{
		state:       state.NewMachineInstanceState(store, machine, eventer, sr),
		runtime:     runtime,
		onDestroyed: onDestroyed,
	}

	return m
}
