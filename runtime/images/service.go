package images

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/containerd/containerd/v2/client"
	ctrderrdefs "github.com/containerd/errdefs"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/registry"

	containerdimages "github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/remotes/docker"
)

type Service struct {
	ctrd        *client.Client
	snapshotter string
}

func NewService(ctrd *client.Client, snapshotter string) *Service {
	return &Service{
		ctrd:        ctrd,
		snapshotter: snapshotter,
	}
}

func (s *Service) Pull(ctx context.Context, ref string, auth registry.RegistriesConfig) (client.Image, error) {
	authorizer := docker.NewDockerAuthorizer(
		docker.WithAuthCreds(func(host string) (string, string, error) {
			return auth.Get(host)
		}),
	)

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(docker.WithClient(newClient()),
			docker.WithAuthorizer(authorizer),
		),
	})

	pullOpts := []client.RemoteOpt{
		client.WithResolver(resolver),
		client.WithPullSnapshotter(s.snapshotter),
		client.WithPullUnpack,
		client.WithChildLabelMap(containerdimages.ChildGCLabelsFilterLayers),
	}

	image, err := s.ctrd.Pull(ctx, ref, pullOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to pull and unpack image %q: %w", ref, err)
	}

	return image, nil
}

func newClient() *http.Client {
	return &http.Client{
		Transport: newTransport(),
	}
}

func newTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:       30 * time.Second,
			KeepAlive:     30 * time.Second,
			FallbackDelay: 300 * time.Millisecond,
		}).DialContext,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
	}
}

func (r *Service) DeleteImage(ctx context.Context, ref string) error {
	slog.Info("Deleting image", "ref", ref)
	err := r.ctrd.ImageService().Delete(ctx, ref)
	if err != nil {
		return fmt.Errorf("failed to delete image %q: %w", ref, err)
	}

	return nil
}

func (r *Service) GetImage(ctx context.Context, ref string) (client.Image, error) {
	i, err := r.ctrd.GetImage(ctx, ref)
	if err != nil {
		if ctrderrdefs.IsNotFound(err) {
			return i, errdefs.NewNotFound(fmt.Sprintf("image %q not found", ref))
		}

		return i, fmt.Errorf("failed to get image %q: %w", ref, err)
	}

	return i, nil

}

func (r *Service) ListImages(ctx context.Context) ([]client.Image, error) {
	return r.ctrd.ListImages(ctx)
}
