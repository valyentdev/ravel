package container

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/runtimes"
)

func (r *Runtime) WaitVM(ctx context.Context, instanceId string) (*runtimes.ExitResult, error) {
	vm, ok := r.runningVMs[instanceId]
	if !ok {
		return nil, core.NewNotFound("instance vm not found")
	}

loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-vm.waitChan:
			break loop
		}

	}

	exitResult := &runtimes.ExitResult{}

	runResult := vm.runResult
	if runResult == nil {
		return exitResult, nil
	}

	exitResult.ExitedAt = runResult.ExitedAt

	exitResult.ExitCode = runResult.ExitCode
	exitResult.Requested = runResult.HasBeenStopped

	if runResult.ExitCode != nil {
		exitResult.Success = *runResult.ExitCode == 0
	}

	return exitResult, nil
}
