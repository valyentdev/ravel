package state

import (
	"context"
	"fmt"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
)

func (s *State) GetAPIMachine(ctx context.Context, namespace, fleetId, id string) (*api.Machine, error) {
	machine, err := s.clusterState.GetAPIMachine(ctx, namespace, fleetId, id)
	if err != nil {
		return nil, err
	}

	return machine, err
}

func (s *State) ListAPIMachines(ctx context.Context, namespace, fleetId string, includeDestroyed bool) ([]api.Machine, error) {
	return s.clusterState.ListAPIMachines(ctx, namespace, fleetId, includeDestroyed)
}

func (s *State) GetMachine(ctx context.Context, namespace, fleetId, id string, showDestroyed bool) (cluster.Machine, error) {
	return s.db.GetMachine(ctx, namespace, fleetId, id, showDestroyed)
}

func (s *State) CreateMachine(machine cluster.Machine, mv api.MachineVersion) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fleet, err := tx.GetFleetForShare(ctx, machine.FleetId)
	if err != nil {
		return fmt.Errorf("failed to get fleet for share: %w", err)
	}

	if fleet.Status != api.FleetStatusActive {
		return errdefs.NewNotFound("fleet not found")
	}

	if err = tx.CreateMachine(ctx, machine, mv); err != nil {
		return fmt.Errorf("failed to create machine on pg: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	err = s.clusterState.CreateMachine(ctx, machine, mv)
	if err != nil {
		return fmt.Errorf("failed to create machine on corro: %w", err)
	}

	return nil
}

func (s *State) UpdateMachine(machine cluster.Machine) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err = tx.UpdateMachine(ctx, machine); err != nil {
		return fmt.Errorf("failed to update machine on pg: %w", err)
	}

	if err = s.clusterState.UpdateMachine(ctx, machine); err != nil {
		return fmt.Errorf("failed to update machine on corro: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (s *State) DestroyMachine(ctx context.Context, id string) error {
	err := s.db.DestroyMachine(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to destroy machine on pg: %w", err)
	}

	err = s.clusterState.DestroyMachine(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to destroy machine on corro: %w", err)
	}

	return nil
}

func (s *State) ListMachineEvents(ctx context.Context, machineId string) ([]api.MachineEvent, error) {
	return s.db.ListMachineEvents(ctx, machineId)
}

func (s *State) ListMachineVersions(ctx context.Context, machineId string) ([]api.MachineVersion, error) {
	return s.db.ListMachineVersions(ctx, machineId)
}

func (s *State) StoreMachineEvent(ctx context.Context, event api.MachineEvent) error {
	return s.db.PushMachineEvent(ctx, event)
}
