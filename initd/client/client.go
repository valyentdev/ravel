package client

import (
	"context"
	"net"
	"net/http"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/internal/httpclient"
	"github.com/valyentdev/ravel/pkg/vsock"
)

type InternalClient struct {
	client *httpclient.Client
}

func NewInternalClient(path string) *InternalClient {
	htc := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return vsock.DialContext(ctx, path, 10000)
			},
			Dial: func(network, addr string) (net.Conn, error) {
				return vsock.Dial(path, 10000)
			},
		},
	}
	c := httpclient.NewClient("http://localhost", htc)

	return newClient(c)
}

func newClient(httpclient *httpclient.Client) *InternalClient {
	return &InternalClient{
		client: httpclient,
	}
}

func (c *InternalClient) Wait(ctx context.Context) (*initd.WaitResult, error) {
	var res initd.WaitResult
	err := c.client.Get(ctx, "/wait", &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *InternalClient) Signal(ctx context.Context, signal int) error {
	req := initd.SignalOptions{
		Signal: signal,
	}
	return c.client.Post(ctx, "/signal", nil, httpclient.WithJSONBody(req))
}

func (c *InternalClient) Exec(ctx context.Context, opts api.ExecOptions) (*api.ExecResult, error) {
	var res api.ExecResult
	err := c.client.Post(ctx, "/exec", &res, httpclient.WithJSONBody(opts))
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *InternalClient) HealthCheck(ctx context.Context) (err error) {
	var res initd.Status
	err = c.client.Get(ctx, "/status", &res)
	if err != nil {
		return err
	}

	if res.Ok {
		return nil
	}

	return errdefs.NewUnknown("status check failed")
}
