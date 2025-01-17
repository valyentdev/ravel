package runtime

import (
	"context"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/daemon"
)

func (r *Runtime) ListImages(ctx context.Context) ([]daemon.Image, error) {
	imagesList, err := r.images.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]images.Image, len(imagesList))

	for i, image := range imagesList {
		result[i] = image.Metadata()
	}

	return result, nil
}

func (r *Runtime) PullImage(ctx context.Context, opt daemon.ImagePullOptions) (*daemon.Image, error) {
	if opt.Ref == "" {
		return nil, errdefs.NewInvalidArgument("ref must be provided")
	}
	ref := opt.Ref
	auth := opt.Auth
	if auth == nil {
		auth = r.registries
	}

	image, err := r.images.Pull(ctx, ref, r.driver.Snapshotter(), auth)
	if err != nil {
		return nil, err
	}

	metadata := image.Metadata()

	return &metadata, nil

}

func (r *Runtime) DeleteImage(ctx context.Context, ref string) error {
	return r.images.DeleteImage(ctx, ref)
}
