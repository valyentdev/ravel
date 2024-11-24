package instancerunner

import (
	"context"

	"github.com/valyentdev/ravel/core/instance"
)

type WaitExitStatus struct {
	err        error
	exitResult instance.ExitResult
	isValid    bool
}

func (w *WaitExitStatus) Err() error {
	return w.err
}

func (w *WaitExitStatus) ExitResult() instance.ExitResult {
	return w.exitResult
}

func (w *WaitExitStatus) IsValid() bool {
	return w.isValid
}

func (ir *InstanceRunner) WaitExit(ctx context.Context) WaitExitStatus {
	ch, ok := ir.waitExit()
	if !ok {
		return WaitExitStatus{isValid: false}
	}

	select {
	case <-ctx.Done():
		return WaitExitStatus{err: ctx.Err()}
	case result := <-ch:
		return WaitExitStatus{exitResult: result, isValid: true}
	}
}

// The channel is closed when the instance cannot have an exit result
// (e.g. it was never started or the result has been lost).
func (ir *InstanceRunner) waitExit() (<-chan instance.ExitResult, bool) {
	ir.lock()
	defer ir.unlock()

	i := ir.Instance()
	status := i.State.Status

	switch status {
	case instance.InstanceStatusRunning:
		ch := make(chan instance.ExitResult, 1)
		ir.waitForExit = append(ir.waitForExit, ch)
		return ch, true
	case instance.InstanceStatusStopped:
		if i.State.ExitResult != nil {
			ch := make(chan instance.ExitResult, 1)
			ch <- *i.State.ExitResult
			return ch, true
		}
		return nil, false
	default:
		return nil, false
	}

}
