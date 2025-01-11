package agentclient

import (
	"context"
	"io"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/internal/httpclient"
	"github.com/valyentdev/ravel/internal/streamutil"
)

func (a *AgentClient) PutMachine(ctx context.Context, opt cluster.PutMachineOptions) (*cluster.MachineInstance, error) {
	var machineInstance cluster.MachineInstance
	err := a.client.Post(ctx, "/machines", &machineInstance, httpclient.WithJSONBody(opt))
	if err != nil {
		return nil, err
	}

	return &machineInstance, nil
}

func (a *AgentClient) StartMachine(ctx context.Context, id string) error {
	err := a.client.Post(ctx, "/machines/"+id+"/start", nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) StopMachine(ctx context.Context, id string, opt *api.StopConfig) error {
	var opts []httpclient.ReqOpt
	if opt != nil {
		opts = append(opts, httpclient.WithJSONBody(opt))
	}

	err := a.client.Post(ctx, "/machines/"+id+"/stop", nil, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) DestroyMachine(ctx context.Context, id string, force bool) error {
	opts := []httpclient.ReqOpt{}
	if force {
		opts = append(opts, httpclient.WithQuery("force", "true"))
	}
	err := a.client.Delete(ctx, "/machines/"+id, opts...)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) MachineExec(ctx context.Context, id string, cmd []string, timeout time.Duration) (*api.ExecResult, error) {
	var result api.ExecResult
	opt := api.ExecOptions{
		Cmd:       cmd,
		TimeoutMs: int(timeout.Milliseconds()),
	}

	err := a.client.Post(ctx, "/instances/"+id+"/exec", &result, httpclient.WithJSONBody(opt))
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *AgentClient) GetMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, error) {
	var logs []*api.LogEntry
	err := a.client.Get(ctx, "/machines/"+id+"/logs", &logs)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

func (a *AgentClient) SubscribeToMachineLogs(ctx context.Context, id string) ([]*api.LogEntry, <-chan *api.LogEntry, error) {
	path := "/machines/" + id + "/logs/follow"

	body, err := a.client.RawGet(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		<-ctx.Done()
		body.Close()
	}()

	return streamutil.SubscribeToLogs(body)
}

func (a *AgentClient) GetMachineLogsRaw(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/machines/" + id + "/logs"
	if follow {
		path += "/follow"
	}
	return a.client.RawGet(ctx, path)
}

func (a *AgentClient) DisableMachineGateway(ctx context.Context, id string) error {
	path := "/machines/" + id + "/gateway/disable"
	err := a.client.Post(ctx, path, nil)
	if err != nil {
		return err
	}
	return nil
}

func (a *AgentClient) EnableMachineGateway(ctx context.Context, id string) error {
	path := "/machines/" + id + "/gateway/enable"
	err := a.client.Post(ctx, path, nil)
	if err != nil {
		return err
	}
	return nil
}
