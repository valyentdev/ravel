package api

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api/httpclient"
)

type AgentClient struct {
	baseUrl    string
	httpClient *http.Client
	client     *httpclient.Client
}

func NewAgentClient(c *http.Client, baseUrl string) *AgentClient {
	return &AgentClient{baseUrl: baseUrl, client: httpclient.NewClient(baseUrl, c), httpClient: c}
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
		opts = append(opts, httpclient.WithQuery("force", "true"))
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

func (a *AgentClient) GetInstanceLogs(ctx context.Context, id string, follow bool) (<-chan *core.LogEntry, error) {
	path := "/instances/" + id + "/logs"

	if follow {
		path += "?follow"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, core.NewUnknown("failed to get instance logs")
	}

	logs := make(chan *core.LogEntry)

	go func() {
		defer close(logs)
		defer resp.Body.Close()
		reader := bufio.NewReader(resp.Body)
		for {
			var log core.LogEntry
			line, _, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					return
				}
				return
			}

			if err := json.Unmarshal(line, &log); err != nil {
				return
			}
			logs <- &log
		}
	}()
	return logs, nil
}

func (a *AgentClient) GetInstanceLogsRaw(ctx context.Context, id string, follow bool) (io.ReadCloser, error) {
	path := "/instances/" + id + "/logs"

	if follow {
		path += "?follow"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, core.NewUnknown("failed to get instance logs")
	}

	return resp.Body, nil
}
