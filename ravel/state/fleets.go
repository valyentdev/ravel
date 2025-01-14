package state

import (
	"context"
	"strings"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
)

func (s *State) CreateFleet(ctx context.Context, fleet api.Fleet) error {
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.GetNamespaceForShare(ctx, fleet.Namespace)
	if err != nil {
		return err
	}

	err = tx.CreateFleet(ctx, fleet)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *State) ListFleets(ctx context.Context, namespace string) ([]api.Fleet, error) {
	fleets, err := s.db.ListFleets(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return fleets, nil
}

func (s *State) GetFleet(ctx context.Context, namespace string, idOrName string) (*api.Fleet, error) {
	if strings.HasPrefix(idOrName, "fleet_") {
		fleet, err := s.db.GetFleetByID(ctx, namespace, idOrName)
		if err != nil {
			return nil, err
		}

		return fleet, nil
	}
	fleet, err := s.db.GetFleetByName(ctx, namespace, idOrName)
	if err != nil {
		return nil, err
	}

	return fleet, nil
}

func (s *State) DestroyFleet(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fleet, err := tx.GetFleetForUpdate(ctx, id)
	if err != nil {
		return err
	}

	if fleet.Status == api.FleetStatusDestroyed {
		return errdefs.NewNotFound("fleet not found")
	}

	machinesCount, err := tx.CountExistingMachinesInFleet(ctx, fleet.Id)
	if err != nil {

		return err
	}
	if machinesCount > 0 {
		return errdefs.NewFailedPrecondition("fleet still has machines")
	}

	err = tx.DeleteFleetGateways(ctx, fleet.Id)
	if err != nil {
		return err
	}

	cstx, err := s.clusterState.BeginTx(context.Background())
	if err != nil {
		return err
	}

	err = cstx.DeleteFleetGateways(ctx, fleet.Id)
	if err != nil {
		return err
	}

	err = tx.DestroyFleet(ctx, fleet.Id)
	if err != nil {
		return err
	}

	err = cstx.Commit(ctx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
