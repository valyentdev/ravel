package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/containerd/containerd/v2/client"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/runtime/disks"
	"github.com/valyentdev/ravel/runtime/instancerunner"
)

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

func (r *Runtime) newInstanceManager(i instance.Instance, disks []disks.Disk) *instancerunner.InstanceRunner {
	return instancerunner.New(r.instancesStore, i, r.driver, disks)
}

func (r *Runtime) CreateInstance(ctx context.Context, opt instance.InstanceOptions) (*instance.Instance, error) {
	id := opt.Id
	var err error
	ok := r.instances.ReserveId(id)
	if !ok {
		err = errdefs.NewAlreadyExists("instance id already in use")
		return nil, err
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

	disksId := make([]string, len(opt.Config.Mounts))
	for i, m := range opt.Config.Mounts {
		disksId[i] = m.Disk
	}

	disks, err := r.disks.GetDisks(disksId...)

	if err = r.disks.AttachInstance(id, disksId...); err != nil {
		return nil, fmt.Errorf("failed to attach disks: %w", err)
	}
	defer func() {
		if err != nil {
			err := r.disks.DetachInstance(disksId...)
			if err != nil {
				slog.Error("failed to detach disks", "err", err)
			}
		}
	}()

	i := instance.Instance{
		Id:       id,
		Metadata: opt.Metadata,
		Config:   opt.Config,
		ImageRef: image.Name(),
		Network:  opt.Network,
		State: instance.State{
			Status: instance.InstanceStatusCreated,
		},
		CreatedAt: time.Now(),
	}

	if err = r.instancesStore.PutInstance(i); err != nil {
		return nil, fmt.Errorf("failed to save instance: %w", err)
	}

	manager := r.newInstanceManager(i, disks)

	r.instances.AddInstance(id, manager)

	return &i, nil
}
