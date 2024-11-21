package agentclient

import (
	"context"

	"github.com/containerd/containerd/v2/core/images"
	"github.com/valyentdev/ravel/runtime"
)

// DeleteImage implements structs.Agent.
func (a *AgentClient) DeleteImage(ctx context.Context, ref string) error {
	return a.client.Delete(ctx, "/images/"+ref)
}

// ListImages implements structs.Agent.
func (a *AgentClient) ListImages(ctx context.Context) ([]images.Image, error) {
	var imagesList []images.Image
	err := a.client.Get(ctx, "/images", &imagesList)
	return imagesList, err
}

// PullImage implements structs.Agent.
func (a *AgentClient) PullImage(ctx context.Context, opts runtime.PullImageOptions) (*images.Image, error) {
	var image images.Image
	err := a.client.Post(ctx, "/images/pull", opts, &image)
	return &image, err
}
