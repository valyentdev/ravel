package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/internal/httpclient"
	"github.com/valyentdev/ravel/internal/streamutil"
)

type DaemonClient struct {
	client *httpclient.Client
}

var _ daemon.Daemon = (*DaemonClient)(nil)

func NewDaemonClient(socket string) *DaemonClient {
	return &DaemonClient{client: httpclient.NewClient("http://localhost", &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socket)
			},
		},
	})}
}

func (a *DaemonClient) CreateInstance(ctx context.Context, options daemon.InstanceOptions) (*instance.Instance, error) {
	var instance instance.Instance
	err := a.client.Post(ctx, "/instances", nil, httpclient.WithJSONBody(&options))
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *DaemonClient) DestroyInstance(ctx context.Context, id string) error {
	err := a.client.Delete(ctx, "/instances/"+id)
	if err != nil {
		return err
	}
	return err
}

func (a *DaemonClient) GetInstance(ctx context.Context, id string) (*instance.Instance, error) {
	var instance instance.Instance
	err := a.client.Get(ctx, "/instances/"+id, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *DaemonClient) InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	opt := api.ExecOptions{
		Cmd:       cmd,
		TimeoutMs: int(timeout.Milliseconds()),
	}
	var result api.ExecResult
	err := a.client.Post(ctx, "/instances/"+id+"/exec", result, httpclient.WithJSONBody(&opt))
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *DaemonClient) ListInstances(ctx context.Context) ([]instance.Instance, error) {
	var instances []instance.Instance
	err := a.client.Get(ctx, "/instances", &instances)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (a *DaemonClient) StartInstance(ctx context.Context, id string) error {
	err := a.client.Post(ctx, "/instances/"+id+"/start", nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *DaemonClient) StopInstance(ctx context.Context, id string, opt *api.StopConfig) error {
	err := a.client.Post(ctx, "/instances/"+id+"/stop", nil, httpclient.WithJSONBody(opt))
	if err != nil {
		return err
	}

	return nil
}

func (a *DaemonClient) GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	var logs []*api.LogEntry
	err := a.client.Get(ctx, "/instances/"+id+"/logs", &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (a *DaemonClient) GetInstanceLogsRaw(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/instances/" + id + "/logs"
	if follow {
		path += "/follow"
	}
	return a.client.RawGet(ctx, path)
}

func (a *DaemonClient) SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	body, err := a.GetInstanceLogsRaw(ctx, id, true)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		<-ctx.Done()
		body.Close()
	}()

	return streamutil.SubscribeToLogs(body)
}

func (c *DaemonClient) DeleteImage(ctx context.Context, ref string) error {
	return c.client.Delete(ctx, fmt.Sprintf("/images/%s", url.PathEscape(ref)))
}

func (c *DaemonClient) ListImages(ctx context.Context) ([]daemon.Image, error) {
	var imagesList []daemon.Image
	err := c.client.Get(ctx, "/images", &imagesList)
	return imagesList, err
}

func (c *DaemonClient) PullImage(ctx context.Context, opts daemon.ImagePullOptions) (*daemon.Image, error) {
	var image daemon.Image
	err := c.client.Post(ctx, "/images/pull", &image, httpclient.WithJSONBody(&opts))
	return &image, err
}
