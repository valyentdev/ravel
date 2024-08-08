package container

import (
	"context"

	"github.com/valyentdev/ravel/internal/agent/runtimes"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

func (r *Runtime) WaitVM(ctx context.Context, instanceId string) (*runtimes.ExitResult, error) {
	vm, ok := r.runningVMs[instanceId]
	if !ok {
		return nil, ravelerrors.ErrInstanceNotFound
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

	exitResult.ExitCode = runResult.ExitCode
	exitResult.Requested = runResult.HasBeenStopped

	if runResult.ExitCode != nil {
		exitResult.Success = *runResult.ExitCode == 0
	}

	return exitResult, nil
}
