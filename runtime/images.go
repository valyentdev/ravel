package runtime

import (
	"context"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/registry"
)

type Image = images.Image

func (r *Runtime) ListImages(ctx context.Context) ([]Image, error) {
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

type PullImageOptions struct {
	Ref  string                       `json:"ref"`
	Auth *registry.RegistryAuthConfig `json:"auth,omitempty"`
}

func (r *Runtime) PullImage(ctx context.Context, opt PullImageOptions) (*Image, error) {
	if opt.Ref == "" {
		return nil, errdefs.NewInvalidArgument("ref must be provided")
	}
	ref := opt.Ref
	auth := opt.Auth
	image, err := r.images.Pull(ctx, ref, auth)
	if err != nil {
		return nil, err
	}

	metadata := image.Metadata()

	return &metadata, nil

}

func (r *Runtime) DeleteImage(ctx context.Context, ref string) error {
	return r.images.DeleteImage(ctx, ref)
}
