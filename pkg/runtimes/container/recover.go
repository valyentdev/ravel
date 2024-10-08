package container

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

func (r *Runtime) RecoverVM(ctx context.Context, instance core.Instance) (runtimes.Handle, bool) {
	h := runtimes.Handle{}
	instanceId := instance.Id
	runningInstance, err := recoverRunningVM(instance)
	if err != nil {
		return h, false
	}

	h.Console = runningInstance.console

	r.runningVMs[instanceId] = runningInstance

	go func() {
		runningInstance.Run()
	}()

	return h, true
}
