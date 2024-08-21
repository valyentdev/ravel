package container

import (
	"context"
	"fmt"

	"github.com/valyentdev/ravel/pkg/core"
)

func (r *Runtime) StartVM(ctx context.Context, instance core.Instance) error {
	_, ok := r.runningVMs[instance.Id]
	if ok {
		return fmt.Errorf("instance %q is already running", instance.Id)
	}

	runningInstance, err := r.createRunningInstance(instance)
	if err != nil {
		return fmt.Errorf("failed to create instance for machine %q: %w", instance.Id, err)
	}

	err = runningInstance.Start()
	if err != nil {
		r.cleanupAfterVMRun(runningInstance)
		return fmt.Errorf("failed to start instance for machine %q: %w", instance.Id, err)
	}

	r.runningVMs[instance.Id] = runningInstance

	go runningInstance.Run()
	return nil
}
