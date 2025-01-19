package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/containerd/containerd/v2/client"
	ctrderr "github.com/containerd/errdefs"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/registry"
	"github.com/valyentdev/ravel/runtime/disks"
	"github.com/valyentdev/ravel/runtime/drivers"
	"github.com/valyentdev/ravel/runtime/drivers/vm"
	"github.com/valyentdev/ravel/runtime/images"
)

func createRavelCgroup() error {
	return os.MkdirAll("/sys/fs/cgroup/ravel", 0755)
}

type Runtime struct {
	instancesStore instance.InstanceStore
	imagesUsage    *images.ImagesUsage
	images         *images.Service
	networking     *networkService
	driver         drivers.Driver
	instances      *State
	disks          *disks.Service
	registries     registry.RegistriesConfig
}

type Store interface {
	instance.InstanceStore
	disks.Store
}

var _ daemon.Runtime = (*Runtime)(nil)

func New(config *config.RuntimeConfig, registries registry.RegistriesConfig, store Store) (*Runtime, error) {
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

	instanceBuilder, err := vm.NewDriver(vm.Config{
		CloudHypervisorBinary: config.CloudHypervisorBinary,
		JailerBinary:          config.JailerBinary,
		InitBinary:            config.InitBinary,
		LinuxKernel:           config.LinuxKernel,
	}, ctrd)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance builder: %w", err)
	}

	runtime := &Runtime{
		disks:          disks.NewService(store, disks.NewZFSPool(config.ZFSPool)),
		instancesStore: store,
		imagesUsage:    imageUsage,
		images:         imagesService,
		networking:     newNetworkService(),
		driver:         instanceBuilder,
		instances:      state,
		registries:     registries,
	}
	return runtime, nil
}

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
		err := r.networking.Allocate(i.Network)
		if err != nil {
			slog.Error("Failed to allocate network for instance %s: %v", i.Id, err)
		}

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
