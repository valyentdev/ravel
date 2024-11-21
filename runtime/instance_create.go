package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	instancemanager "github.com/valyentdev/ravel/runtime/manager"
)

type InstanceOptions struct {
	Metadata instance.InstanceMetadata `json:"metadata"`
	Config   instance.InstanceConfig   `json:"config"`
}

func (r *Runtime) PruneImages(ctx context.Context) error {
	return nil
}

func (r *Runtime) useImage(ref string) (client.Image, error) {
	r.imagesUsage.Lock()
	defer r.imagesUsage.Unlock()
	image, err := r.images.GetImage(context.Background(), ref)
	if err != nil {
		return image, err
	}

	r.imagesUsage.UseImage(image.Name())
	return image, nil
}

func (r *Runtime) releaseImage(ref string) {
	r.imagesUsage.Lock()
	r.imagesUsage.ReleaseImage(ref)
	r.imagesUsage.Unlock()
}

func (r *Runtime) newInstanceManager(i instance.Instance) *instancemanager.Manager {
	return instancemanager.NewInstanceManager(r.instancesStore, i, r.networking, r.instanceBuilder, r.eventReporter)
}

func (r *Runtime) CreateInstance(ctx context.Context, id string, opt InstanceOptions) (*instance.Instance, error) {
	network, err := r.networking.AllocateNext()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate network resources: %w", err)
	}

	defer func() {
		if err != nil {
			r.networking.Release(network)
		}
	}()

	ok := r.instances.ReserveId(id)
	if !ok {
		return nil, errdefs.NewAlreadyExists("instance id already in use")
	}
	defer func() {
		if err != nil {
			r.instances.ReleaseId(id)
		}
	}()

	image, err := r.useImage(opt.Config.Image)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			r.releaseImage(opt.Config.Image)
		}
	}()

	i := instance.Instance{
		Id:       id,
		Metadata: opt.Metadata,
		Config:   opt.Config,
		ImageRef: image.Name(),
		Network:  network,
		State: instance.State{
			Status: instance.InstanceStatusStopped,
		},
		CreatedAt: time.Now(),
	}

	if err := r.instanceBuilder.PrepareInstance(ctx, &i, image); err != nil {
		return nil, fmt.Errorf("failed to prepare instance: %w", err)
	}

	defer func() {
		if err != nil {
			r.instanceBuilder.CleanupInstance(ctx, &i)
		}
	}()

	if err := r.instancesStore.PutInstance(i); err != nil {
		return nil, fmt.Errorf("failed to save instance: %w", err)
	}

	manager := r.newInstanceManager(i)

	r.instances.AddInstance(id, manager)

	return &i, nil
}
