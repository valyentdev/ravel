package ravel

import (
	"context"
	"strings"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/internal/id"
)

func (r *Ravel) CreateFleet(ctx context.Context, ns string, name string) (*api.Fleet, error) {
	if err := validateObjectName(name); err != nil {
		return nil, errdefs.NewInvalidArgument(err.Error())
	}

	namespace, err := r.GetNamespace(ctx, ns)
	if err != nil {
		return nil, err
	}

	fleet := api.Fleet{
		Id:        id.GeneratePrefixed("fleet"),
		Namespace: namespace.Name,
		Name:      name,
		CreatedAt: time.Now(),
	}

	err = r.db.CreateFleet(ctx, fleet)
	if err != nil {
		return nil, err
	}

	return &fleet, nil
}

func (r *Ravel) ListFleets(ctx context.Context, namespace string) ([]api.Fleet, error) {
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	fleets, err := r.db.ListFleets(ctx, namespace)
	if err != nil {
		return nil, err
	}

	return fleets, nil
}

func (r *Ravel) GetFleet(ctx context.Context, namespace string, idOrName string) (*api.Fleet, error) {
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(idOrName, "fleet_") {
		fleet, err := r.db.GetFleetByID(ctx, namespace, idOrName)
		if err != nil {
			return nil, err
		}

		return fleet, nil
	}
	fleet, err := r.db.GetFleetByName(ctx, namespace, idOrName)
	if err != nil {
		return nil, err
	}

	return fleet, nil
}

func (r *Ravel) DestroyFleet(ctx context.Context, namespace string, id string) error {
	fleet, err := r.GetFleet(ctx, namespace, id)
	if err != nil {
		return err
	}

	err = r.db.DestroyFleet(ctx, fleet.Namespace, fleet.Id)
	if err != nil {
		return err
	}

	return nil
}
