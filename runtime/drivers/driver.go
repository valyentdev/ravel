package drivers

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/runtime/disks"
)

type InstanceTask interface {
	Start(ctx context.Context) error
	// StartFromSnapshot starts the VM by restoring from a snapshot instead of cold booting
	// globalSnapshotPath is the source on the host, jailSnapshotPath is the jail-relative destination
	StartFromSnapshot(ctx context.Context, globalSnapshotPath, jailSnapshotPath string) error
	Exec(ctx context.Context, cmd []string, timeout time.Duration) (*api.ExecResult, error)
	Run() instance.ExitResult
	WaitExit(ctx context.Context) bool
	Signal(ctx context.Context, signal string) error
	Stop(ctx context.Context, signal string) error
	Shutdown(ctx context.Context) error
	// Snapshot saves the VM state to a file for fast restore
	Snapshot(ctx context.Context, path string) error
	// Restore restores the VM state from a snapshot file
	Restore(ctx context.Context, path string) error
}

type Driver interface {
	BuildInstanceTask(ctx context.Context, instance *instance.Instance, disks []disks.Disk) (InstanceTask, error)
	CleanupInstanceTask(ctx context.Context, instance *instance.Instance) error
	RecoverInstanceTask(ctx context.Context, i *instance.Instance) (InstanceTask, error)
	CleanupInstance(ctx context.Context, instance *instance.Instance) error
	Snapshotter() string
}
