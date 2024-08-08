package container

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

func (r *Runtime) RecoverVM(ctx context.Context, instance core.Instance) bool {
	instanceId := instance.Id
	runningInstance, err := recoverRunningVM(instance)
	if err != nil {
		return false
	}

	r.runningVMs[instanceId] = runningInstance

	go func() {
		runningInstance.Run()
	}()

	return true
}
