package images

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/containerd/containerd/v2/client"
	ctrderrdefs "github.com/containerd/errdefs"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/core/registry"

	containerdimages "github.com/containerd/containerd/v2/core/images"
	"github.com/containerd/containerd/v2/core/remotes/docker"
	distribution "github.com/distribution/reference"
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

func (s *Service) Pull(ctx context.Context, name string, auth *registry.RegistryAuthConfig) (client.Image, error) {
	namedRef, err := distribution.ParseDockerRef(name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %q: %w", name, err)
	}
	ref := namedRef.String()

	authorizer := docker.NewDockerAuthorizer(
		docker.WithAuthCreds(func(host string) (string, string, error) {
			return parseAuth(auth, host)
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

// extracted from github.com/containerd/containerd/internal/cri/server/images/image_pull.go
func parseAuth(auth *registry.RegistryAuthConfig, host string) (string, string, error) {
	if auth == nil {
		return "", "", nil
	}
	if auth.ServerAddress != "" {
		// Do not return the auth info when server address doesn't match.
		u, err := url.Parse(auth.ServerAddress)
		if err != nil {
			return "", "", fmt.Errorf("parse server address: %w", err)
		}
		if host != u.Host {
			return "", "", nil
		}
	}
	if auth.Username != "" {
		return auth.Username, auth.Password, nil
	}
	if auth.IdentityToken != "" {
		return "", auth.IdentityToken, nil
	}
	if auth.Auth != "" {
		decLen := base64.StdEncoding.DecodedLen(len(auth.Auth))
		decoded := make([]byte, decLen)
		_, err := base64.StdEncoding.Decode(decoded, []byte(auth.Auth))
		if err != nil {
			return "", "", err
		}
		user, passwd, ok := strings.Cut(string(decoded), ":")
		if !ok {
			return "", "", fmt.Errorf("invalid decoded auth: %q", decoded)
		}
		return user, strings.Trim(passwd, "\x00"), nil
	}
	return "", "", nil
}

func (r *Service) DeleteImage(ctx context.Context, ref string) error {
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
