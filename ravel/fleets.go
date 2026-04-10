package ravel

import (
	"context"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/internal/id"
)

func (r *Ravel) CreateFleet(ctx context.Context, ns string, name string, metadata *api.Metadata) (*api.Fleet, error) {
	if err := validateObjectName(name); err != nil {
		return nil, errdefs.NewInvalidArgument(err.Error())
	}

	// Validate metadata if provided
	if err := ValidateMetadata(metadata); err != nil {
		return nil, err
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
		Status:    api.FleetStatusActive,
		Metadata:  metadata,
	}

	err = r.State.CreateFleet(ctx, fleet)
	if err != nil {
		return nil, err
	}

	return &fleet, nil
}

func (r *Ravel) ListFleets(ctx context.Context, namespace string, labelFilters map[string]string) ([]api.Fleet, error) {
	_, err := r.GetNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	fleets, err := r.State.ListFleets(ctx, namespace, labelFilters)
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

	return r.State.GetFleet(ctx, namespace, idOrName)
}

func (r *Ravel) DestroyFleet(ctx context.Context, namespace string, idOrName string) error {
	fleet, err := r.GetFleet(ctx, namespace, idOrName)
	if err != nil {
		return err
	}

	if err := r.State.DestroyFleet(ctx, fleet.Id); err != nil {
		return err
	}

	return nil
}

// UpdateFleetMetadata updates the metadata for a fleet
func (r *Ravel) UpdateFleetMetadata(ctx context.Context, namespace string, idOrName string, metadata api.Metadata) (*api.Fleet, error) {
	// Validate metadata
	if err := ValidateMetadata(&metadata); err != nil {
		return nil, err
	}

	// Get the fleet to ensure it exists and to get its ID
	fleet, err := r.GetFleet(ctx, namespace, idOrName)
	if err != nil {
		return nil, err
	}

	// Update the metadata in the state layer
	if err := r.State.UpdateFleetMetadata(ctx, fleet.Id, metadata); err != nil {
		return nil, err
	}

	// Return the updated fleet
	return r.GetFleet(ctx, namespace, fleet.Id)
}
