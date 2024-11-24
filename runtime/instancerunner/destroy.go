package instancerunner

import (
	"context"
	"fmt"

	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (ir *InstanceRunner) Destroy(ctx context.Context) error {
	ir.lock()
	defer ir.unlock()
	return ir.destroyImpl(ctx)
}

// destroyImpl destroys the instance.
// It MUST be called with the lock held.
func (ir *InstanceRunner) destroyImpl(ctx context.Context) error {
	i := ir.Instance()

	status := ir.Status()

	if status != instance.InstanceStatusStopped {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s state", status))
	}

	err := ir.updateInstanceState(instance.State{
		Status: instance.InstanceStatusDestroying,
	})
	if err != nil {
		return err
	}

	err = ir.vmBuilder.CleanupInstance(ctx, &i)
	if err != nil {
		return err
	}

	return nil
}
