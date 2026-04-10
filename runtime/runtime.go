// Package runtime provides the core Ravel Runtime implementation for managing
// microVM instances. It handles OCI image management via containerd, VM lifecycle
// through Cloud Hypervisor, disk management, and instance state tracking.
//
// The Runtime is responsible for:
//   - Running OCI images inside cloud-hypervisor microVMs
//   - Managing VM lifecycle (create, start, stop, destroy)
//   - Image management through containerd with devmapper snapshotter
//   - Disk provisioning and management via ZFS
//   - Instance state persistence and recovery
package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/alexisbouchez/ravel/core/config"
	"github.com/alexisbouchez/ravel/core/daemon"
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/alexisbouchez/ravel/core/registry"
	"github.com/alexisbouchez/ravel/runtime/disks"
	"github.com/alexisbouchez/ravel/runtime/drivers"
	"github.com/alexisbouchez/ravel/runtime/drivers/firecracker"
	"github.com/alexisbouchez/ravel/runtime/drivers/vm"
	"github.com/alexisbouchez/ravel/runtime/images"
	"github.com/containerd/containerd/v2/client"
	ctrderr "github.com/containerd/errdefs"
)

// createRavelCgroup creates the root cgroup for all Ravel instances.
// This cgroup is used to manage resource limits for all microVMs.
func createRavelCgroup() error {
	return os.MkdirAll("/sys/fs/cgroup/ravel", 0755)
}

// Runtime is the core component responsible for managing microVM instances.
// It orchestrates containerd for image management, Cloud Hypervisor for VM
// execution, and ZFS for disk provisioning.
type Runtime struct {
	instancesStore instance.InstanceStore    // Persistent storage for instance metadata
	imagesUsage    *images.ImagesUsage       // Tracks which images are in use
	images         *images.Service           // OCI image management service
	driver         drivers.Driver            // VM driver (Cloud Hypervisor)
	instances      *State                    // In-memory state of running instances
	disks          *disks.Service            // Disk provisioning and management
	registries     registry.RegistriesConfig // OCI registry configurations
}

// Store combines instance and disk storage interfaces required by the Runtime.
type Store interface {
	instance.InstanceStore
	disks.Store
}

// Ensure Runtime implements the daemon.Runtime interface.
var _ daemon.Runtime = (*Runtime)(nil)

// New creates and initializes a new Runtime instance with the provided configuration.
// It sets up:
//   - Containerd client connection
//   - VM driver (CloudHypervisor or Firecracker based on config)
//   - Image management service
//   - Disk service with ZFS backend
//
// Returns an error if any initialization step fails.
func New(runtimeConfig *config.RuntimeConfig, registries registry.RegistriesConfig, store Store) (*Runtime, error) {
	err := os.MkdirAll("/var/lib/ravel/instances", 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances directory: %w", err)
	}

	ctrd, err := initContainerd()
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	imagesService := images.NewService(ctrd)
	imageUsage := images.NewImagesUsage()

	state := NewState()

	// Select driver based on runtime type configuration
	var driver drivers.Driver
	runtimeType := runtimeConfig.GetRuntimeType()

	switch runtimeType {
	case config.RuntimeTypeFirecracker:
		slog.Info("Using Firecracker runtime")
		driver, err = firecracker.NewDriver(firecracker.Config{
			FirecrackerBinary: runtimeConfig.FirecrackerBinary,
			JailerBinary:      runtimeConfig.JailerBinary,
			InitBinary:        runtimeConfig.InitBinary,
			LinuxKernel:       runtimeConfig.LinuxKernel,
		}, ctrd)
	default: // CloudHypervisor (default)
		slog.Info("Using CloudHypervisor runtime")
		driver, err = vm.NewDriver(vm.Config{
			CloudHypervisorBinary: runtimeConfig.CloudHypervisorBinary,
			JailerBinary:          runtimeConfig.JailerBinary,
			InitBinary:            runtimeConfig.InitBinary,
			LinuxKernel:           runtimeConfig.LinuxKernel,
		}, ctrd)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create VM driver: %w", err)
	}

	runtime := &Runtime{
		disks:          disks.NewService(store, disks.NewZFSPool(runtimeConfig.ZFSPool)),
		instancesStore: store,
		imagesUsage:    imageUsage,
		images:         imagesService,
		driver:         driver,
		instances:      state,
		registries:     registries,
	}
	return runtime, nil
}

// initContainerd initializes a connection to the containerd daemon and ensures
// the "ravel" namespace exists. This namespace isolates Ravel's images and
// containers from other containerd users on the system.
func initContainerd() (*client.Client, error) {
	ctrd, err := client.New("/var/run/containerd/containerd.sock", client.WithDefaultNamespace("ravel"))
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	err = ctrd.NamespaceService().Create(context.Background(), "ravel", map[string]string{})
	if err != nil {
		if ctrderr.IsAlreadyExists(err) {
			slog.Info("Namespace already exists")
			return ctrd, nil
		}
		return nil, fmt.Errorf("failed to create containerd namespace: %w", err)
	}

	return ctrd, nil
}

// Start initializes the Runtime and recovers any existing instances from persistent
// storage. It:
//   - Creates the root Ravel cgroup
//   - Loads all instances from the store
//   - Recovers each instance's state (reattaches to running VMs)
//   - Provisions disks for each instance
//
// This method should be called once during daemon startup.
func (r *Runtime) Start() error {
	slog.Info("Starting runtime")
	err := createRavelCgroup()
	if err != nil {
		return fmt.Errorf("failed to create ravel cgroup: %w", err)
	}

	instances, err := r.instancesStore.LoadInstances()
	if err != nil {
		return fmt.Errorf("failed to load instances: %w", err)
	}

	for _, i := range instances {
		disks, err := r.disks.GetDisks(i.Config.GetDisks()...)
		if err != nil {
			slog.Error("Failed to get disks for instance %s: %v", i.Id, err)
			return err
		}

		r.imagesUsage.UseImage(i.ImageRef)
		manager := r.newInstanceManager(i, disks)
		manager.Recover()
		r.instances.AddInstance(i.Id, manager)
	}

	slog.Info("Runtime started")

	return nil
}
