package instancerunner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/core/instance"
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

// StartFromSnapshot starts the instance by restoring from a snapshot instead of cold booting.
// This enables sub-100ms cold starts for AI sandbox workloads.
// globalSnapshotPath is the path to the global snapshot storage
// jailSnapshotPath is the jail-relative path where the snapshot will be placed
func (ir *InstanceRunner) StartFromSnapshot(ctx context.Context, globalSnapshotPath, jailSnapshotPath string) error {
	ir.lock()
	defer ir.unlock()
	slog.Debug("starting instance from snapshot", "id", ir.Instance().Id, "globalPath", globalSnapshotPath, "jailPath", jailSnapshotPath)

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

	err = runner.StartFromSnapshot(globalSnapshotPath, jailSnapshotPath)
	if err != nil {
		return err
	}

	ir.updateInstanceState(instance.State{Status: instance.InstanceStatusRunning})

	go ir.run()

	return err
}
