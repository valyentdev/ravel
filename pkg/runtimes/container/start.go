package container

import (
	"context"
	"fmt"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

func (r *Runtime) StartVM(ctx context.Context, instance core.Instance) (runtimes.Handle, error) {
	h := runtimes.Handle{}
	_, ok := r.runningVMs[instance.Id]
	if ok {
		return h, fmt.Errorf("instance %q is already running", instance.Id)
	}

	runningInstance, err := r.createRunningInstance(instance)
	if err != nil {
		return h, fmt.Errorf("failed to create instance for machine %q: %w", instance.Id, err)
	}

	h, err = runningInstance.Start()
	if err != nil {
		r.cleanupAfterVMRun(runningInstance)
		return h, fmt.Errorf("failed to start instance for machine %q: %w", instance.Id, err)
	}

	r.runningVMs[instance.Id] = runningInstance

	go runningInstance.Run()
	return h, nil
}
