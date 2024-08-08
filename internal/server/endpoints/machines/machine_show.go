package machines

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

type GetMachineRequest struct {
	Namespace string `header:"namespace"`
	FleetID   string `path:"fleet_id"`
	MachineID string `path:"machine_id"`
}

type GetMachineResponse struct {
	Body *core.Machine `json:"body"`
}

func (s *Endpoint) getMachine(ctx context.Context, req *GetMachineRequest) (*GetMachineResponse, error) {
	machine, err := s.m.GetMachine(ctx, req.Namespace, req.FleetID, req.MachineID)
	if err != nil {
		return nil, err
	}

	return &GetMachineResponse{
		Body: &machine,
	}, nil
}
