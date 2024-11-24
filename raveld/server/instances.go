package server

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/internal/streamutil"
)

type CreateInstanceRequest struct {
	Body daemon.InstanceOptions
}

type CreateInstanceResponse struct {
	Body *instance.Instance
}

func (s *DaemonServer) createInstance(ctx context.Context, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	res, err := s.daemon.CreateInstance(ctx, req.Body)
	if err != nil {
		s.log("Failed to create instance", err)
		return nil, err
	}
	return &CreateInstanceResponse{Body: res}, nil
}

type ListInstancesRequest struct {
}

type ListInstancesResponse struct {
	Body []instance.Instance
}

func (s *DaemonServer) listInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	res, err := s.daemon.ListInstances(ctx)
	if err != nil {
		s.log("Failed to list instances", err)
		return nil, err
	}
	return &ListInstancesResponse{Body: res}, nil
}

type GetInstanceRequest struct {
	Id string `path:"id"`
}

type GetInstanceResponse struct {
	Body instance.Instance
}

func (s *DaemonServer) getInstance(ctx context.Context, req *GetInstanceRequest) (*GetInstanceResponse, error) {
	res, err := s.daemon.GetInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance", err)
		return nil, err
	}
	return &GetInstanceResponse{Body: *res}, nil
}

type DestroyInstanceRequest struct {
	Id string `path:"id"`
}

type DestroyInstanceResponse struct {
}

func (s *DaemonServer) destroyInstance(ctx context.Context, req *DestroyInstanceRequest) (*DestroyInstanceResponse, error) {
	err := s.daemon.DestroyInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to destroy instance", err)
		return nil, err
	}
	return &DestroyInstanceResponse{}, nil
}

type StartInstanceRequest struct {
	Id string `path:"id"`
}

type StartInstanceResponse struct {
}

func (s *DaemonServer) startInstance(ctx context.Context, req *StartInstanceRequest) (*StartInstanceResponse, error) {
	err := s.daemon.StartInstance(ctx, req.Id)
	if err != nil {
		s.log("Failed to start instance", err)
		return nil, err
	}
	return &StartInstanceResponse{}, nil
}

type StopInstanceRequest struct {
	Id   string          `path:"id"`
	Body *api.StopConfig `required:"false"`
}

type StopInstanceResponse struct {
}

func (s *DaemonServer) stopInstance(ctx context.Context, req *StopInstanceRequest) (*StopInstanceResponse, error) {
	err := s.daemon.StopInstance(ctx, req.Id, req.Body)
	if err != nil {
		s.log("Failed to stop instance", err)
		return nil, err
	}
	return &StopInstanceResponse{}, nil
}

type ExecBody struct {
	Cmd     []string `json:"cmd"`
	Timeout int      `json:"timeout"`
}

type ExecRequest struct {
	Id   string `path:"id"`
	Body api.ExecOptions
}

type ExecResponse struct {
	Body *api.ExecResult
}

func (s *DaemonServer) exec(ctx context.Context, req *ExecRequest) (*ExecResponse, error) {
	res, err := s.daemon.InstanceExec(ctx, req.Id, req.Body.Cmd, time.Duration(req.Body.TimeoutMs)*time.Millisecond)
	if err != nil {
		s.log("Failed to exec command", err)
		return nil, err
	}
	return &ExecResponse{Body: res}, nil
}

type FollowInstanceLogsRequest struct {
	Id string `path:"id"`
}

func (s *DaemonServer) followInstanceLogs(ctx context.Context, req *GetInstanceLogsRequest) (*huma.StreamResponse, error) {
	logs, logsChan, err := s.daemon.SubscribeToInstanceLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance logs", err)
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			streamutil.StreamLogs(ctx, logs, logsChan)
		}}, nil
}

type GetInstanceLogsRequest struct {
	Id     string `path:"id"`
	Follow bool   `query:"follow"`
}

type GetInstanceLogsResponse struct {
	Body []*api.LogEntry
}

func (s *DaemonServer) getInstanceLogs(ctx context.Context, req *GetInstanceLogsRequest) (*GetInstanceLogsResponse, error) {
	logs, err := s.daemon.GetInstanceLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to get instance logs", err)
		return nil, err
	}

	return &GetInstanceLogsResponse{Body: logs}, nil
}
