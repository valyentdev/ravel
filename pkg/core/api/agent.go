package api

import (
	"context"
	"net/http"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api/httpclient"
)

type AgentClient struct {
	client *httpclient.Client
}

var _ core.Agent = (*AgentClient)(nil)

func NewAgentClient(c *http.Client, baseUrl string) *AgentClient {
	return &AgentClient{client: httpclient.NewClient(baseUrl, c)}
}

func (a *AgentClient) CreateInstance(ctx context.Context, options core.CreateInstancePayload) (*core.Instance, error) {
	var instance core.Instance
	err := a.client.Post(ctx, "/instances", options, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *AgentClient) DestroyInstance(ctx context.Context, id string, force bool) error {
	var opts []httpclient.ReqOpt
	if force {
		opts = append(opts, httpclient.WithQuery("force", ""))
	}
	err := a.client.Delete(ctx, "/instances/"+id, opts...)
	if err != nil {
		return err
	}
	return err
}

func (a *AgentClient) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	var instance core.Instance
	err := a.client.Get(ctx, "/instances/"+id, &instance)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (a *AgentClient) InstanceExec(ctx context.Context, id string, opt core.InstanceExecOptions) (*core.ExecResult, error) {
	var result core.ExecResult
	err := a.client.Post(ctx, "/instances/"+id+"/exec", opt, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (a *AgentClient) ListInstances(ctx context.Context) ([]core.Instance, error) {
	var instances []core.Instance
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

func (a *AgentClient) StopInstance(ctx context.Context, id string, opt *core.StopConfig) error {
	err := a.client.Post(ctx, "/instances/"+id+"/stop", opt, nil)
	if err != nil {
		return err
	}

	return nil
}
