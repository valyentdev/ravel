package container

import (
	"errors"
	"fmt"
	"sync"

	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/internal/agent/runtimes"
	"github.com/valyentdev/ravel/internal/agent/runtimes/container/containerd"
	"github.com/valyentdev/ravel/internal/agent/runtimes/container/filesystems"
	"github.com/valyentdev/ravel/internal/agent/runtimes/container/image"

	"github.com/valyentdev/ravel/pkg/config"
)

type Runtime struct {
	mutex            sync.RWMutex
	config           config.AgentConfig
	containerdClient *client.Client
	fsBuilder        *filesystems.ContainerFSBuilder
	images           *image.ImagesService
	runningVMs       map[string]*runningVM
}

var _ runtimes.Runtime = (*Runtime)(nil)

func NewRuntime(
	config config.AgentConfig,
) (runtimes.Runtime, error) {

	containerdClient, err := containerd.NewContainerdClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	images := image.NewImagesService(containerdClient)

	fsBuilder, err := filesystems.NewContainerFSBuilder(containerdClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem builder: %w", err)
	}

	return &Runtime{
		config:           config,
		containerdClient: containerdClient,
		fsBuilder:        fsBuilder,
		images:           images,
		runningVMs:       make(map[string]*runningVM),
	}, nil
}

func (r *Runtime) Stop() error {
	errs := []error{}

	err := r.containerdClient.Close()
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to close containerd client: %w", err))
	}

	return errors.Join(errs...)
}
