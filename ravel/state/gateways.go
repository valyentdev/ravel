package state

import (
	"context"
	"fmt"
	"strings"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

func (s *State) GetGateway(ctx context.Context, namespace, fleetId, idOrName string) (api.Gateway, error) {
	if strings.HasPrefix(idOrName, "gw_") {
		return s.db.GetGatewayById(ctx, namespace, fleetId, idOrName)
	}

	return s.db.GetGatewayByName(ctx, namespace, fleetId, idOrName)

}

func (s *State) CreateGateway(gateway api.Gateway) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fleet, err := tx.GetFleetForShare(ctx, gateway.FleetId)
	if err != nil {
		return fmt.Errorf("failed to get fleet for share: %w", err)
	}

	if fleet.Status != api.FleetStatusActive {
		return errdefs.NewNotFound("fleet not found")
	}

	if err = tx.CreateGateway(ctx, gateway); err != nil {
		return fmt.Errorf("failed to create gateway on pg: %w", err)
	}

	cstx, err := s.clusterState.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	err = cstx.CreateGateway(ctx, gateway)
	if err != nil {
		return fmt.Errorf("failed to create gateway on corro: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if err = cstx.Commit(ctx); err != nil { // commit in corrosion after commit in pg to prevent inconsistency
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (s *State) DeleteGateway(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err = tx.DeleteGateway(ctx, id); err != nil {
		return fmt.Errorf("failed to delete gateway on pg: %w", err)
	}

	cstx, err := s.clusterState.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	err = cstx.DeleteGateway(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete gateway on corro: %w", err)
	}

	if err = cstx.Commit(ctx); err != nil { // commit in corrosion first to prevent untracked gateways in corrosion in case of failure
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	return nil
}

func (s *State) ListGateways(ctx context.Context, namespace string) ([]api.Gateway, error) {
	return s.db.ListGateways(ctx, namespace)
}

func (s *State) ListGatewaysOnFleet(ctx context.Context, ns, fleetId string) ([]api.Gateway, error) {
	return s.db.ListGatewaysOnFleet(ctx, ns, fleetId)
}
