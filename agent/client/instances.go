package agentclient

import (
	"context"
	"io"
	"time"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/internal/httpclient"
)

func (a *AgentClient) CreateInstance(ctx context.Context, options structs.InstanceOptions) (*instance.Instance, error) {
	var instance instance.Instance
	err := a.client.Post(ctx, "/instances", options, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *AgentClient) DestroyInstance(ctx context.Context, id string, force bool) error {
	var opts []httpclient.ReqOpt
	if force {
		opts = append(opts, httpclient.WithQuery("force", "true"))
	}
	err := a.client.Delete(ctx, "/instances/"+id, opts...)
	if err != nil {
		return err
	}
	return err
}

func (a *AgentClient) GetInstance(ctx context.Context, id string) (*instance.Instance, error) {
	var instance instance.Instance
	err := a.client.Get(ctx, "/instances/"+id, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *AgentClient) InstanceExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	opt := api.ExecOptions{
		Cmd:       cmd,
		TimeoutMs: int(timeout.Milliseconds()),
	}
	var result api.ExecResult
	err := a.client.Post(ctx, "/instances/"+id+"/exec", opt, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *AgentClient) ListInstances(ctx context.Context) ([]instance.Instance, error) {
	var instances []instance.Instance
	err := a.client.Get(ctx, "/instances", &instances)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (a *AgentClient) StartInstance(ctx context.Context, id string) error {
	err := a.client.Post(ctx, "/instances/"+id+"/start", nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) StopInstance(ctx context.Context, id string, opt *api.StopConfig) error {
	err := a.client.Post(ctx, "/instances/"+id+"/stop", opt, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) GetInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	var logs []*api.LogEntry
	err := a.client.Get(ctx, "/instances/"+id+"/logs", &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (a *AgentClient) GetInstanceLogsRaw(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/instances/" + id + "/logs"
	if follow {
		path += "/follow"
	}
	return a.getLogsRaw(ctx, path)
}

func (a *AgentClient) SubscribeToInstanceLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	path := "/instances/" + id + "/logs/follow"
	body, err := a.getLogsRaw(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	go func() {
		<-ctx.Done()
		body.Close()
	}()

	return subscribeToLogs(body)
}
