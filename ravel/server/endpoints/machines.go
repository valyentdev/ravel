package endpoints

import (
	"bufio"
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

type CreateMachineBody = api.CreateMachinePayload
type CreateMachineRequest struct {
	FleetResolver
	Body *CreateMachineBody
}

type CreateMachineResponse struct {
	Body *api.Machine
}

func (s *Endpoints) createMachine(ctx context.Context, req *CreateMachineRequest) (*CreateMachineResponse, error) {
	m, err := s.ravel.CreateMachine(ctx, req.Namespace, req.Fleet, *req.Body)
	if err != nil {
		s.log("Failed to create machine", err)
		return nil, err
	}

	return &CreateMachineResponse{
		Body: m,
	}, nil

}

type DestroyMachineRequest struct {
	MachineResolver
	Force bool `query:"force"`
}

type DestroyMachineResponse struct {
}

func (e *Endpoints) destroyMachine(ctx context.Context, req *DestroyMachineRequest) (*DestroyMachineResponse, error) {
	err := e.ravel.DestroyMachine(ctx, req.Namespace, req.Fleet, req.MachineId, req.Force)
	if err != nil {
		e.log("Failed to destroy machine", err)
		return nil, err
	}

	return nil, nil
}

type ListMachinesRequest struct {
	FleetResolver
	IncludeDestroyed bool `query:"destroyed"`
}

type ListMachinesResponse struct {
	Body []api.Machine `json:"machines"`
}

func (e *Endpoints) listMachines(ctx context.Context, req *ListMachinesRequest) (*ListMachinesResponse, error) {
	machines, err := e.ravel.ListMachines(ctx, req.Namespace, req.Fleet, req.IncludeDestroyed)
	if err != nil {
		e.log("Failed to list machines", err)
		return nil, err
	}

	return &ListMachinesResponse{
		Body: machines,
	}, nil
}

type GetMachineRequest struct {
	MachineResolver
}

type GetMachineResponse struct {
	Body *api.Machine `json:"body"`
}

func (e *Endpoints) getMachine(ctx context.Context, req *GetMachineRequest) (*GetMachineResponse, error) {
	machine, err := e.ravel.GetMachine(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to get machine", err)
		return nil, err
	}

	return &GetMachineResponse{
		Body: machine,
	}, nil
}

type ListMachineVersionsRequest struct {
	MachineResolver
}

type ListMachineVersionsResponse struct {
	Body []api.MachineVersion `json:"body"`
}

func (e *Endpoints) listMachineVersions(ctx context.Context, req *ListMachineVersionsRequest) (*ListMachineVersionsResponse, error) {
	mvs, err := e.ravel.ListMachineVersions(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to list machine versions", err)
		return nil, err
	}

	return &ListMachineVersionsResponse{
		Body: mvs,
	}, nil
}

type StartMachineRequest struct {
	MachineResolver
}

type StartMachineResponse struct {
}

func (e *Endpoints) startMachine(ctx context.Context, req *StartMachineRequest) (*StartMachineResponse, error) {
	err := e.ravel.StartMachine(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to start machine", err)
		return nil, err
	}

	return nil, nil
}

type StopMachineRequest struct {
	MachineResolver
	Body    *api.StopConfig
	RawBody []byte
}

type StopMachineResponse struct {
}

func (e *Endpoints) stopMachine(ctx context.Context, req *StopMachineRequest) (*StopMachineResponse, error) {
	slog.Info("Stopping machine", "machine_id", req.Body)
	err := e.ravel.StopMachine(ctx, req.Namespace, req.Fleet, req.MachineId, req.Body)
	if err != nil {
		e.log("Failed to stop machine", err)
		return nil, err
	}

	return nil, nil
}

type ExecCmdRequest struct {
	MachineResolver
	Body *api.ExecOptions
}

func (e *Endpoints) machineExec(ctx context.Context, req *ExecCmdRequest) (*api.ExecResult, error) {
	res, err := e.ravel.MachineExec(ctx, req.Namespace, req.Fleet, req.MachineId, req.Body)
	if err != nil {
		e.log("Failed to execute command", err)
		return nil, err
	}

	return res, nil
}

type GetMachineLogsRequest struct {
	MachineResolver
	Follow bool `query:"follow"`
}

type GetMachineLogsResponse struct {
	Body string
}

func (e *Endpoints) getMachineLogs(ctx context.Context, req *GetMachineLogsRequest) (*huma.StreamResponse, error) {
	logs, err := e.ravel.GetMachineLogsRaw(ctx, req.Namespace, req.Fleet, req.MachineId, req.Follow)
	if err != nil {
		e.log("Failed to get machine logs", err)
		return nil, err
	}

	return &huma.StreamResponse{
		Body: func(ctx huma.Context) {
			ctx.AppendHeader("Content-Type", "application/x-ndjson")
			ctx.SetStatus(200)

			bw := ctx.BodyWriter()
			rw := bw.(http.ResponseWriter)
			rc := http.NewResponseController(rw)
			buff := bufio.NewReader(logs)
			for {
				line, err := buff.ReadBytes('\n')
				if err != nil {
					break
				}

				rw.Write(line)
				rc.Flush()
			}

		},
	}, nil
}

type ListMachineEventsRequest struct {
	MachineResolver
}

type ListMachineEventsResponse struct {
	Body []api.MachineEvent
}

func (e *Endpoints) listMachineEvents(ctx context.Context, req *ListMachineEventsRequest) (*ListMachineEventsResponse, error) {
	events, err := e.ravel.ListMachineEvents(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to list machine events", err)
		return nil, err
	}

	return &ListMachineEventsResponse{
		Body: events,
	}, nil
}

type WaitMachineStatusRequest struct {
	MachineResolver
	Timeout uint              `query:"timeout" minimum:"1" maximum:"60"`
	Status  api.MachineStatus `query:"status" required:"true"`
}

type WaitMachineStatusResponse struct {
}

func (e *Endpoints) waitMachineStatus(ctx context.Context, req *WaitMachineStatusRequest) (*WaitMachineStatusResponse, error) {
	if req.Timeout > 60 || req.Timeout < 1 {
		return nil, errdefs.NewInvalidArgument("timeout must be between 1 and 60 seconds")
	}

	err := e.ravel.WaitMachineStatus(ctx, req.Namespace, req.Fleet, req.MachineId, req.Status, req.Timeout)
	if err != nil {
		e.log("Failed to wait for machine status", err)
		return nil, err
	}

	return nil, nil
}

type EnableMachineGatewayRequest struct {
	MachineResolver
}

type EnableMachineGatewayResponse struct {
}

func (e *Endpoints) enableMachineGateway(ctx context.Context, req *EnableMachineGatewayRequest) (*EnableMachineGatewayResponse, error) {
	err := e.ravel.EnableMachineGateway(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to enable machine gateway", err)
		return nil, err
	}

	return nil, nil
}

type DisableMachineGatewayRequest struct {
	MachineResolver
}

type DisableMachineGatewayResponse struct {
}

func (e *Endpoints) disableMachineGateway(ctx context.Context, req *DisableMachineGatewayRequest) (*DisableMachineGatewayResponse, error) {
	err := e.ravel.DisableMachineGateway(ctx, req.Namespace, req.Fleet, req.MachineId)
	if err != nil {
		e.log("Failed to disable machine gateway", err)
		return nil, err
	}

	return nil, nil
}
