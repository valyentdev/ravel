package runtime

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/containerd/containerd/v2/client"
	ctrderr "github.com/containerd/errdefs"
	"github.com/valyentdev/ravel/core/images"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/vm"
)

type Runtime struct {
	instancesStore  instance.InstanceStore
	eventReporter   instance.EventReporter
	imagesUsage     *images.ImagesUsage
	images          *images.Service
	networking      *networkService
	instanceBuilder instance.Builder
	instances       *State
}

func New(is instance.InstanceStore, es instance.EventReporter, initBinary string, linuxKernel string) (*Runtime, error) {
	ctrd, err := initContainerd()
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	imagesService := images.NewService(ctrd, "devmapper")
	imageUsage := images.NewImagesUsage()

	state := NewState()

	instanceBuilder := vm.NewBuilder("/var/run/ravel", "/var/lib/ravel", initBinary, linuxKernel, imagesService, ctrd, "devmapper")

	runtime := &Runtime{
		instancesStore:  is,
		eventReporter:   es,
		imagesUsage:     imageUsage,
		images:          imagesService,
		networking:      newNetworkService(),
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
	instances, err := r.instancesStore.LoadInstances()
	if err != nil {
		return fmt.Errorf("failed to load instances: %w", err)
	}

	for _, i := range instances {
		r.networking.Allocate(i.Network)
		r.imagesUsage.UseImage(i.ImageRef)
		manager := r.newInstanceManager(i)
		manager.Recover()
		r.instances.AddInstance(i.Id, manager)
	}

	slog.Info("Runtime started")

	return nil
}
