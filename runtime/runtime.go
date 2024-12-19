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
	"github.com/valyentdev/ravel/internal/resources"
	"github.com/valyentdev/ravel/runtime/images"
	"github.com/valyentdev/ravel/runtime/vm"
)

func createRavelCgroup() error {
	return os.MkdirAll("/sys/fs/cgroup/ravel", 0755)
}

type Runtime struct {
	instancesStore  instance.InstanceStore
	imagesUsage     *images.ImagesUsage
	images          *images.Service
	networking      *networkService
	instanceBuilder instance.Builder
	instances       *State
	registries      registry.RegistriesConfig
}

var _ daemon.Runtime = (*Runtime)(nil)

func New(config *config.RuntimeConfig, registries registry.RegistriesConfig, is instance.InstanceStore) (*Runtime, error) {
	err := os.MkdirAll("/var/lib/ravel/instances", 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create instances directory: %w", err)
	}

	uid, gid, err := setupRavelJailerUser()
	if err != nil {
		return nil, fmt.Errorf("failed to setup ravel jailer user: %w", err)
	}

	frequency, err := resources.GetHostCPUFrequency()
	if err != nil {
		return nil, fmt.Errorf("failed to get host CPU frequency: %w", err)
	}

	slog.Info("Host CPU frequency", "mhz", frequency)

	ctrd, err := initContainerd()
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	const snapshotter = "devmapper"
	imagesService := images.NewService(ctrd, snapshotter)
	imageUsage := images.NewImagesUsage()

	state := NewState()

	initBinary := config.InitBinary
	linuxKernel := config.LinuxKernel
	jailer := config.JailerBinary

	instanceBuilder, err := vm.NewBuilder(config.CloudHypervisorBinary, jailer, initBinary, linuxKernel, imagesService, ctrd, snapshotter, frequency, vm.User{Uid: uid, Gid: gid})
	if err != nil {
		return nil, fmt.Errorf("failed to create instance builder: %w", err)
	}

	runtime := &Runtime{
		instancesStore:  is,
		imagesUsage:     imageUsage,
		images:          imagesService,
		networking:      newNetworkService(vm.User{Uid: uid, Gid: gid}),
		instanceBuilder: instanceBuilder,
		instances:       state,
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

		r.imagesUsage.UseImage(i.ImageRef)
		manager := r.newInstanceManager(i)
		manager.Recover()
		r.instances.AddInstance(i.Id, manager)
	}

	slog.Info("Runtime started")

	return nil
}
