package endpoints

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/ravel"
)

type CreateMachineBody = ravel.CreateMachineOptions
type CreateMachineRequest struct {
	FleetResolver
	Body *CreateMachineBody
}

type CreateMachineResponse struct {
	Body *ravel.Machine
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
	Body []ravel.Machine `json:"machines"`
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
	Body *ravel.Machine `json:"body"`
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
	Body []core.MachineVersion `json:"body"`
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
	Body *core.StopConfig
}

type StopMachineResponse struct {
}

func (e *Endpoints) stopMachine(ctx context.Context, req *StopMachineRequest) (*StopMachineResponse, error) {
	err := e.ravel.StopMachine(ctx, req.Namespace, req.Fleet, req.MachineId, req.Body)
	if err != nil {
		e.log("Failed to stop machine", err)
		return nil, err
	}

	return nil, nil
}
