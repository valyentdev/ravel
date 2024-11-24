package instancerunner

import (
	"github.com/valyentdev/ravel/core/instance"
)

func (ir *InstanceRunner) notifyAndClearWaitForExit(result instance.ExitResult) {
	for _, ch := range ir.waitForExit {
		ch <- result
	}
	ir.waitForExit = nil
}

func (ir *InstanceRunner) run() {
	result := ir.runner.Run()
	ir.lock()
	ir.updateInstanceState(instance.State{
		Status:     instance.InstanceStatusStopped,
		ExitResult: &result,
	})
	ir.notifyAndClearWaitForExit(result)
	ir.unlock()
}
