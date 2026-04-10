// Package containerd provides a native containerd runtime driver for Ravel.
// This driver runs containers directly without microVMs for lower overhead
// while still maintaining isolation via Linux namespaces and cgroups.
package containerd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/internal/resources"
	"github.com/alexisbouchez/ravel/runtime/disks"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/containerd/cgroups/v3/cgroup2"
	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/snapshots"
)

const (
	dataDir     = "/var/lib/ravel/containers"
	snapshotter = "overlayfs"
)

// Driver implements the drivers.Driver interface using native containerd.
type Driver struct {
	cpuMhz      int64
	ctrd        *client.Client
	snapshotter snapshots.Snapshotter
}

// Config holds configuration for the containerd driver.
type Config struct {
	// No special config needed - uses containerd defaults
}

// NewDriver creates a new containerd driver instance.
func NewDriver(ctrd *client.Client) (*Driver, error) {
	frequency, err := resources.GetHostCPUFrequency()
	if err != nil {
		return nil, fmt.Errorf("failed to get host CPU frequency: %w", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create root cgroup for ravel containers
	_, err = cgroup2.NewManager("/sys/fs/cgroup", "/ravel-containers", &cgroup2.Resources{})
	if err != nil {
		return nil, fmt.Errorf("failed to create cgroup2 manager: %w", err)
	}

	snapshotService := ctrd.SnapshotService(snapshotter)

	return &Driver{
		cpuMhz:      frequency,
		ctrd:        ctrd,
		snapshotter: snapshotService,
	}, nil
}

var _ drivers.Driver = (*Driver)(nil)

// Snapshotter returns the snapshotter type used by this driver.
func (d *Driver) Snapshotter() string {
	return snapshotter
}

// getCgroupMemory creates memory cgroup configuration.
func getCgroupMemory(c *instance.InstanceGuestConfig) *cgroup2.Memory {
	max := int64(c.MemoryMB * 1_000_000)
	return &cgroup2.Memory{
		Max: &max,
	}
}

// getCgroupCPU creates CPU cgroup configuration.
func getCgroupCPU(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.CPU {
	quota := int64(float64(c.CpusMHz) / float64(cpuMhz) * 1_000_000)
	period := uint64(1_000_000)
	max := cgroup2.NewCPUMax(&quota, &period)
	return &cgroup2.CPU{
		Max: max,
	}
}

// getInstanceResources creates cgroup resource limits.
func getInstanceResources(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.Resources {
	return &cgroup2.Resources{
		CPU:    getCgroupCPU(cpuMhz, c),
		Memory: getCgroupMemory(c),
	}
}

// getInstanceDir returns the working directory for an instance.
func getInstanceDir(id string) string {
	return path.Join(dataDir, id)
}

// BuildInstanceTask creates a new container task for the instance.
func (d *Driver) BuildInstanceTask(ctx context.Context, inst *instance.Instance, disks []disks.Disk) (drivers.InstanceTask, error) {
	instanceDir := getInstanceDir(inst.Id)
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create instance directory: %w", err)
	}

	// Get image
	image, err := d.ctrd.GetImage(ctx, inst.ImageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	// Prepare snapshot for container
	snapshotName := inst.Id
	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rootfs: %w", err)
	}

	// Create snapshot from image layers
	parent := ""
	if len(diffIDs) > 0 {
		parent = diffIDs[len(diffIDs)-1].String()
	}

	_, err = d.snapshotter.Prepare(ctx, snapshotName, parent)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare snapshot: %w", err)
	}

	// Create cgroup manager
	resources := getInstanceResources(d.cpuMhz, &inst.Config.Guest)
	cgroupPath := path.Join("/ravel-containers", inst.Id)
	cgroupManager, err := cgroup2.NewManager("/sys/fs/cgroup", cgroupPath, resources)
	if err != nil {
		d.snapshotter.Remove(ctx, snapshotName)
		return nil, fmt.Errorf("failed to create cgroup manager: %w", err)
	}

	return &containerTask{
		id:            inst.Id,
		snapshotName:  snapshotName,
		image:         image,
		config:        &inst.Config,
		ctrd:          d.ctrd,
		cgroupManager: cgroupManager,
		cgroupPath:    cgroupPath,
		stopConfig:    inst.Config.Stop,
		waitChan:      make(chan struct{}),
	}, nil
}

// CleanupInstanceTask cleans up a stopped container task.
func (d *Driver) CleanupInstanceTask(ctx context.Context, inst *instance.Instance) error {
	// Remove snapshot
	if err := d.snapshotter.Remove(ctx, inst.Id); err != nil {
		slog.Warn("snapshot not found during cleanup", "id", inst.Id, "error", err)
	}

	// Remove instance directory
	instanceDir := getInstanceDir(inst.Id)
	if err := os.RemoveAll(instanceDir); err != nil {
		slog.Error("failed to remove instance directory", "error", err)
	}

	return nil
}

// RecoverInstanceTask attempts to recover a running container.
func (d *Driver) RecoverInstanceTask(ctx context.Context, inst *instance.Instance) (drivers.InstanceTask, error) {
	// For now, containerd driver doesn't support recovery
	// Containers are ephemeral and should be restarted on daemon restart
	return nil, fmt.Errorf("containerd driver does not support task recovery")
}

// CleanupInstance performs final cleanup for a destroyed instance.
func (d *Driver) CleanupInstance(ctx context.Context, inst *instance.Instance) error {
	instanceDir := getInstanceDir(inst.Id)
	if err := os.RemoveAll(instanceDir); err != nil {
		return fmt.Errorf("failed to remove instance directory: %w", err)
	}

	// Cleanup cgroup
	cgroupPath := path.Join("/ravel-containers", inst.Id)
	cgroupManager, err := cgroup2.Load(cgroupPath)
	if err == nil {
		if err := cgroupManager.Delete(); err != nil {
			slog.Error("failed to delete cgroup", "error", err)
		}
	}

	return nil
}
