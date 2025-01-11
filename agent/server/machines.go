package server

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/internal/streamutil"
)

type CreateMachineRequest struct {
	Body cluster.PutMachineOptions
}

type CreateMachineResponse struct {
	Body cluster.MachineInstance
}

func (s *AgentServer) putMachine(ctx context.Context, req *CreateMachineRequest) (*CreateMachineResponse, error) {
	mi, err := s.agent.PutMachine(ctx, req.Body)
	if err != nil {
		s.log("Failed to put machine", err)
		return nil, err
	}

	return &CreateMachineResponse{Body: *mi}, nil

}

type DestroyMachineRequest struct {
	Id    string `path:"id"`
	Force bool   `query:"force"`
}

type DestroyMachineResponse struct {
}

func (s *AgentServer) destroyMachine(ctx context.Context, req *DestroyMachineRequest) (*DestroyMachineResponse, error) {
	err := s.agent.DestroyMachine(ctx, req.Id, req.Force)
	if err != nil {
		s.log("Failed to delete machine", err)
		return nil, err
	}

	return &DestroyMachineResponse{}, nil
}

type MachineExecRequest struct {
	Id   string `path:"id"`
	Body api.ExecOptions
}

type MachineExecResponse struct {
	Body *api.ExecResult
}

func (s *AgentServer) machineExec(ctx context.Context, req *MachineExecRequest) (*MachineExecResponse, error) {
	res, err := s.agent.MachineExec(ctx, req.Id, req.Body.Cmd, req.Body.GetTimeout())
	if err != nil {
		s.log("Failed to exec machine", err)
		return nil, err
	}
	return &MachineExecResponse{Body: res}, nil
}

type StartMachineRequest struct {
	Id string `path:"id"`
}

type StartMachineResponse struct {
}

func (s *AgentServer) startMachine(ctx context.Context, req *StartMachineRequest) (*StartMachineResponse, error) {
	err := s.agent.StartMachine(ctx, req.Id)
	if err != nil {
		s.log("Failed to start machine", err)
		return nil, err
	}
	return &StartMachineResponse{}, nil
}

type StopMachineRequest struct {
	Id   string `path:"id"`
	Body *api.StopConfig
}

type StopMachineResponse struct {
}

func (s *AgentServer) stopMachine(ctx context.Context, req *StopMachineRequest) (*StopMachineResponse, error) {
	err := s.agent.StopMachine(ctx, req.Id, req.Body)
	if err != nil {
		s.log("Failed to stop machine", err)
		return nil, err
	}
	return &StopMachineResponse{}, nil
}

type FollowMachineLogsRequest struct {
	Id string `path:"id"`
}

func (s *AgentServer) followMachineLogs(ctx context.Context, req *FollowMachineLogsRequest) (*huma.StreamResponse, error) {
	logs, ch, err := s.agent.SubscribeToMachineLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to follow machine logs", err)
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			streamutil.StreamLogs(ctx, logs, ch)
		}}, nil
}

type GetMachineLogsRequest struct {
	Id string `path:"id"`
}

type GetMachineLogsResponse struct {
	Body []*api.LogEntry
}

func (s *AgentServer) getMachineLogs(ctx context.Context, req *GetMachineLogsRequest) (*GetMachineLogsResponse, error) {
	logs, err := s.agent.GetMachineLogs(ctx, req.Id)
	if err != nil {
		s.log("Failed to get machine logs", err)
		return nil, err
	}
	return &GetMachineLogsResponse{Body: logs}, nil
}

type EnableMachineGatewayRequest struct {
	Id string `path:"id"`
}

type EnableMachineGatewayResponse struct {
}

func (s *AgentServer) enableMachineGateway(ctx context.Context, req *EnableMachineGatewayRequest) (*EnableMachineGatewayResponse, error) {
	err := s.agent.EnableMachineGateway(ctx, req.Id)
	if err != nil {
		s.log("Failed to enable machine gateway", err)
		return nil, err
	}
	return &EnableMachineGatewayResponse{}, nil
}

type DisableMachineGatewayRequest struct {
	Id string `path:"id"`
}

type DisableMachineGatewayResponse struct {
}

func (s *AgentServer) disableMachineGateway(ctx context.Context, req *DisableMachineGatewayRequest) (*DisableMachineGatewayResponse, error) {
	err := s.agent.DisableMachineGateway(ctx, req.Id)
	if err != nil {
		s.log("Failed to disable machine gateway", err)
		return nil, err
	}
	return &DisableMachineGatewayResponse{}, nil
}
