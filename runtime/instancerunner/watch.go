package instancerunner

import (
	"context"

	"github.com/valyentdev/ravel/core/instance"
)

func (ir *InstanceRunner) WatchState(ctx context.Context) <-chan instance.State {
	sub := ir.stateObserver.Subscribe()

	ch := sub.Ch()

	go func() {
		<-ctx.Done()
		sub.Unsubscribe()
	}()

	return ch
}
