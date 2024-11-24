package instancerunner

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/core/instance"
)

func (ir *InstanceRunner) Recover() {
	ir.lock()
	defer ir.unlock()
	state := ir.Instance().State
	status := state.Status

	// All these cases potentially involve a still running VM that needs to be recovered.
	if status == instance.InstanceStatusRunning || status == instance.InstanceStatusStarting || state.Stopping {
		ir.recoverRunning(state)
		return
	}

	if status == instance.InstanceStatusDestroying {
		ir.recoverDestroying()
		return
	}
}

func (ir *InstanceRunner) recoverRunning(state instance.State) {
	status := state.Status
	runner := ir.newVMRunner()
	err := runner.Recover()
	if err != nil {
		requested := state.Stopping
		ir.updateInstanceState(instance.State{
			Status: instance.InstanceStatusStopped,
			ExitResult: &instance.ExitResult{
				ExitCode:  -1,
				Requested: requested,
				ExitedAt:  time.Now(),
			},
		})
		return
	}

	ir.setVMRunner(runner)

	if status != instance.InstanceStatusRunning {
		ir.updateInstanceState(instance.State{
			Status: instance.InstanceStatusRunning,
		})
	}

	go ir.run()
}

func (ir *InstanceRunner) recoverDestroying() {
	err := ir.destroyImpl(context.Background())
	if err != nil {
		slog.Error("failed to destroy instance", "error", err)
	}
}
