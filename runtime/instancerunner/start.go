package instancerunner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
)

func (ir *InstanceRunner) Start(ctx context.Context) error {
	ir.lock()
	defer ir.unlock()
	slog.Debug("starting instance", "id", ir.Instance().Id)

	if ir.Status() != instance.InstanceStatusStopped && ir.Status() != instance.InstanceStatusCreated {
		return errdefs.NewFailedPrecondition(fmt.Sprintf("instance is in %s status", ir.Status()))
	}

	err := ir.updateInstanceState(instance.State{
		Status: instance.InstanceStatusStarting,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			ir.updateInstanceState(instance.State{Status: instance.InstanceStatusStopped})
		}
	}()

	runner := ir.newVMRunner()
	ir.setVMRunner(runner)

	err = runner.Start()
	if err != nil {
		return err
	}

	ir.updateInstanceState(instance.State{Status: instance.InstanceStatusRunning})

	go ir.run()

	return err
}
