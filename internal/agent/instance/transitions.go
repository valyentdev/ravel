package instance

import (
	"github.com/valyentdev/ravel/pkg/core"
)

var possiblesTransitions = map[core.MachineStatus]map[core.MachineStatus]struct{}{
	core.MachineStatusCreated: {
		core.MachineStatusPreparing: {},
	},
	core.MachineStatusPreparing: {
		core.MachineStatusStopped:    {},
		core.MachineStatusDestroying: {},
	},

	core.MachineStatusStopped: {
		core.MachineStatusStarting:   {},
		core.MachineStatusDestroying: {},
	},

	core.MachineStatusStarting: {
		core.MachineStatusRunning: {},
		core.MachineStatusStopped: {},
	},

	core.MachineStatusRunning: {
		core.MachineStatusStopping: {},
		core.MachineStatusStopped:  {},
	},

	core.MachineStatusStopping: {
		core.MachineStatusStopped: {},
	},

	core.MachineStatusDestroying: {
		core.MachineStatusDestroyed: {},
	},

	core.MachineStatusDestroyed: {},
}

func canTransition(from core.MachineStatus, to core.MachineStatus) bool {
	possiblesNext, ok := possiblesTransitions[from]
	if !ok {
		return false
	}
	_, ok = possiblesNext[to]
	return ok
}
