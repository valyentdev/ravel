package machines

import (
	"context"
)

type DestroyMachineRequest struct {
	FleetID   string `path:"fleet_id"`
	MachineID string `path:"machine_id"`
	Namespace string `header:"namespace"`
}

type DestroyMachineResponse struct {
}

func (s *Endpoint) destroyMachine(ctx context.Context, req *DestroyMachineRequest) (*DestroyMachineResponse, error) {
	err := s.m.DestroyMachine(ctx, req.Namespace, req.FleetID, req.MachineID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
