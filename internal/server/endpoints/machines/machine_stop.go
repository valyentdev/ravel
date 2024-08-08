package machines

import (
	"context"
)

type StopMachineRequest struct {
	FleetID   string `path:"fleet_id"`
	MachineID string `path:"machine_id"`
	Namespace string `header:"namespace"`
}

type StopMachineResponse struct {
}

func (s *Endpoint) stopMachine(ctx context.Context, req *StopMachineRequest) (*StopMachineResponse, error) {
	err := s.m.StopMachine(ctx, req.Namespace, req.FleetID, req.MachineID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
