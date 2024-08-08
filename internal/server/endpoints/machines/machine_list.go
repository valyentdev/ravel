package machines

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

type ListMachinesRequest struct {
	Namespace        string `header:"namespace"`
	FleetNameOrID    string `path:"fleet_id"`
	IncludeDestroyed bool   `query:"destroyed"`
}

type ListMachinesResponse struct {
	Body []core.Machine `json:"machines"`
}

func (s *Endpoint) listMachines(ctx context.Context, req *ListMachinesRequest) (*ListMachinesResponse, error) {
	machines, err := s.m.ListMachines(ctx, req.Namespace, req.FleetNameOrID, req.IncludeDestroyed)
	if err != nil {
		return nil, err
	}

	return &ListMachinesResponse{
		Body: machines,
	}, nil
}
