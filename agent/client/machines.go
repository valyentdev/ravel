package agentclient

import (
	"context"
	"io"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
)

func (a *AgentClient) PutMachine(ctx context.Context, opt cluster.PutMachineOptions) (*cluster.MachineInstance, error) {
	var machineInstance cluster.MachineInstance
	err := a.client.Post(ctx, "/machines", opt, &machineInstance)
	if err != nil {
		return nil, err
	}

	return &machineInstance, nil
}

func (a *AgentClient) StartMachine(ctx context.Context, id string) error {
	err := a.client.Post(ctx, "/machines/"+id+"/start", nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) StopMachine(ctx context.Context, id string, opt *api.StopConfig) error {
	err := a.client.Post(ctx, "/machines/"+id+"/stop", opt, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *AgentClient) DestroyMachine(ctx context.Context, id string, force bool) error {
	err := a.client.Delete(ctx, "/machines/"+id)
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

	err := a.client.Post(ctx, "/instances/"+id+"/exec", opt, &result)
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

func (a *AgentClient) GetMachineLogsRaw(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/machines/" + id + "/logs"
	if follow {
		path += "/follow"
	}
	return a.getLogsRaw(ctx, path)
}
