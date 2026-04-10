package containerd

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/containerd/v2/client"
)

// containerTask implements the drivers.InstanceTask interface for containerd containers.
// Note: This is a simplified implementation that runs containers via containerd,
// providing lower overhead than microVMs but less isolation.
type containerTask struct {
	id            string
	snapshotName  string
	image         client.Image
	config        *instance.InstanceConfig
	ctrd          *client.Client
	container     client.Container
	task          client.Task
	cgroupManager *cgroup2.Manager
	cgroupPath    string
	stopConfig    *api.StopConfig
	exitStatus    <-chan client.ExitStatus
	waitChan      chan struct{}
	exitResult    *instance.ExitResult
	stopRequested bool
}

var _ drivers.InstanceTask = (*containerTask)(nil)

// Start starts the container task.
func (ct *containerTask) Start(ctx context.Context) error {
	// For containerd driver, we don't use high-level Container API
	// Instead we work directly with containerd primitives
	// This is a simplified placeholder - a full implementation would:
	// 1. Create OCI runtime spec
	// 2. Use containerd's runtime to start container
	// 3. Monitor via containerd events

	// For now, return an error indicating this needs proper implementation
	return fmt.Errorf("containerd driver start not fully implemented - needs OCI runtime integration")
}

// StartFromSnapshot starts the container by restoring from a snapshot.
// Note: Containerd containers don't support VM-style snapshots.
func (ct *containerTask) StartFromSnapshot(ctx context.Context, globalSnapshotPath, jailSnapshotPath string) error {
	return fmt.Errorf("containerd driver does not support snapshot restore")
}

// Snapshot saves the container state to a file.
// Note: Containerd containers don't support VM-style snapshots.
func (ct *containerTask) Snapshot(ctx context.Context, path string) error {
	return fmt.Errorf("containerd driver does not support snapshots")
}

// Restore restores the container state from a snapshot file.
// Note: Containerd containers don't support VM-style snapshots.
func (ct *containerTask) Restore(ctx context.Context, path string) error {
	return fmt.Errorf("containerd driver does not support snapshot restore")
}

// monitor watches the task and handles exit.
func (ct *containerTask) monitor() {
	defer close(ct.waitChan)

	exitStatus := <-ct.exitStatus
	ct.exitResult = &instance.ExitResult{
		Success:   exitStatus.ExitCode() == 0,
		ExitCode:  int(exitStatus.ExitCode()),
		ExitedAt:  exitStatus.ExitTime(),
		Requested: ct.stopRequested,
	}

	slog.Debug("container exited", "id", ct.id, "exitCode", exitStatus.ExitCode())

	// Cleanup task
	if ct.task != nil {
		if _, err := ct.task.Delete(context.Background()); err != nil {
			slog.Error("failed to delete task", "id", ct.id, "error", err)
		}
	}
}

// Stop stops the container with the configured signal.
func (ct *containerTask) Stop(ctx context.Context, signal string) error {
	ct.stopRequested = true
	return fmt.Errorf("not implemented")
}

// Signal sends a signal to the container process.
func (ct *containerTask) Signal(ctx context.Context, signal string) error {
	return fmt.Errorf("not implemented")
}

// Shutdown forcefully stops the container.
func (ct *containerTask) Shutdown(ctx context.Context) error {
	return nil // Cleanup handled in CleanupInstanceTask
}

// Run waits for the container to exit and returns the result.
func (ct *containerTask) Run() instance.ExitResult {
	<-ct.waitChan
	if ct.exitResult == nil {
		return instance.ExitResult{
			Success:  false,
			ExitCode: -1,
			ExitedAt: time.Now(),
		}
	}
	return *ct.exitResult
}

// WaitExit waits for the container to exit.
func (ct *containerTask) WaitExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-ct.waitChan:
		return true
	}
}

// Exec executes a command inside the running container.
func (ct *containerTask) Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	return nil, fmt.Errorf("not implemented")
}

// parseSignal converts signal string to syscall.Signal.
func parseSignal(signal string) syscall.Signal {
	switch signal {
	case "SIGTERM":
		return syscall.SIGTERM
	case "SIGKILL":
		return syscall.SIGKILL
	case "SIGINT":
		return syscall.SIGINT
	case "SIGHUP":
		return syscall.SIGHUP
	default:
		return syscall.SIGTERM
	}
}
