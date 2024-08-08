package machines

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/manager"
)

type CreateMachineBody = manager.CreateMachineOptions
type CreateMachineRequest struct {
	FleetID   string `path:"fleet_id"`
	Namespace string `header:"namespace"`
	Body      *CreateMachineBody
}

type CreateMachineResponse struct {
	Body *core.Machine
}

func (s *Endpoint) createMachine(ctx context.Context, req *CreateMachineRequest) (*CreateMachineResponse, error) {
	err := s.m.CreateMachine(ctx, req.Namespace, req.FleetID, *req.Body)
	if err != nil {
		return nil, err
	}

	return &CreateMachineResponse{
		Body: nil,
	}, nil

}
