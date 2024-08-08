package image

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/containerd/containerd/v2/client"
	containerdimages "github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	distribution "github.com/distribution/reference"
	"github.com/valyentdev/ravel/pkg/core"
)

type ImagesService struct {
	client *client.Client
}

func NewImagesService(client *client.Client) *ImagesService {
	return &ImagesService{client: client}
}

func (i *ImagesService) Pull(ctx context.Context, name string, auth *core.RegistryAuthConfig) (client.Image, error) {
	namedRef, err := distribution.ParseDockerRef(name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %q: %w", name, err)
	}
	ref := namedRef.String()

	authorizer := docker.NewDockerAuthorizer(
		docker.WithAuthCreds(func(host string) (string, string, error) {
			return ParseAuth(auth, host)
		}),
	)

	resolver := docker.NewResolver(docker.ResolverOptions{
		Hosts: docker.ConfigureDefaultRegistries(docker.WithClient(newClient()),
			docker.WithAuthorizer(authorizer),
		),
	})

	pullOpts := []client.RemoteOpt{
		client.WithResolver(resolver),
		client.WithPullSnapshotter("devmapper"),
		client.WithPullUnpack,
		client.WithChildLabelMap(containerdimages.ChildGCLabelsFilterLayers),
	}

	image, err := i.client.Pull(ctx, ref, pullOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to pull and unpack image %q: %w", ref, err)
	}

	return image, nil
}

func (i *ImagesService) GetImage(ctx context.Context, ref string) (client.Image, error) {
	return i.client.GetImage(ctx, ref)
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
