package machines

import (
	"context"
)

type StartMachineRequest struct {
	FleetID   string `path:"fleet_id"`
	MachineID string `path:"machine_id"`
	Namespace string `header:"namespace"`
}

type StartMachineResponse struct {
}

func (s *Endpoint) startMachine(ctx context.Context, req *StartMachineRequest) (*StartMachineResponse, error) {
	err := s.m.StartMachine(ctx, req.Namespace, req.FleetID, req.MachineID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
