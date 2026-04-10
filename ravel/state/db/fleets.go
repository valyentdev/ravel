package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/internal/dbutil"
	"github.com/alexisbouchez/ravel/ravel/state/db/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func scanFleet(row dbutil.Scannable) (*api.Fleet, error) {
	var fleet api.Fleet
	var metadataJSON []byte
	err := row.Scan(&fleet.Id, &fleet.Namespace, &fleet.Name, &fleet.CreatedAt, &fleet.Status, &metadataJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errdefs.NewNotFound("fleet not found")
		}

		return nil, err
	}

	// Deserialize metadata if present
	if len(metadataJSON) > 0 && string(metadataJSON) != "{}" {
		var metadata api.Metadata
		if err := json.Unmarshal(metadataJSON, &metadata); err == nil {
			// Only set metadata if it has labels or annotations
			if len(metadata.Labels) > 0 || len(metadata.Annotations) > 0 {
				fleet.Metadata = &metadata
			}
		}
	}

	return &fleet, nil
}

func (q *Queries) CreateFleet(ctx context.Context, fleet api.Fleet) error {
	var metadataJSON []byte
	var err error

	if fleet.Metadata != nil {
		metadataJSON, err = json.Marshal(fleet.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	_, err = q.db.Exec(ctx, `INSERT INTO fleets (id, namespace, name, created_at, metadata) VALUES ($1, $2, $3, $4, $5)`, fleet.Id, fleet.Namespace, fleet.Name, fleet.CreatedAt, metadataJSON)
	if err != nil {
		var pg *pgconn.PgError
		if errors.As(err, &pg) {
			if pg.ConstraintName == schema.UniqueFleetNameConstraint {
				return errdefs.NewAlreadyExists("fleet name already exists in namespace")
			}
			if pg.ConstraintName == "fleets_namespace_fkey" {
				return errdefs.NewNotFound("namespace not found")
			}
		}
		return err
	}

	return nil
}

func (q *Queries) ListFleets(ctx context.Context, namespace string, labelFilters map[string]string) ([]api.Fleet, error) {
	query := `SELECT id, namespace, name, created_at, status, metadata FROM fleets WHERE namespace = $1 AND status = 'active'`
	args := []interface{}{namespace}

	// Add label filtering if provided
	if len(labelFilters) > 0 {
		labelsJSON, err := json.Marshal(labelFilters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label filters: %w", err)
		}
		query += ` AND metadata->'labels' @> $2::jsonb`
		args = append(args, labelsJSON)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	fleets := []api.Fleet{}
	for rows.Next() {
		fleet, err := scanFleet(rows)
		if err != nil {
			return nil, err
		}

		fleets = append(fleets, *fleet)
	}
	return fleets, nil
}

func (q *Queries) CountActiveFleets(ctx context.Context, namespace string) (int, error) {
	var count int
	err := q.db.QueryRow(ctx, `SELECT COUNT(*) FROM fleets WHERE namespace = $1 AND status = 'active'`, namespace).Scan(&count)
	return count, err
}

func (q *Queries) GetFleetByName(ctx context.Context, namespace string, name string) (*api.Fleet, error) {
	return q.getFleet(ctx, `namespace = $1 AND name = $2 AND status = 'active'`, namespace, name)
}

func (q *Queries) GetFleetForUpdate(ctx context.Context, id string) (*api.Fleet, error) {
	return q.getFleet(ctx, `id = $1 FOR UPDATE`, id)
}

func (q *Queries) GetFleetForShare(ctx context.Context, id string) (*api.Fleet, error) {
	return q.getFleet(ctx, `id = $1 FOR SHARE`, id)
}

func (q *Queries) GetFleetByID(ctx context.Context, namespace string, id string) (*api.Fleet, error) {
	return q.getFleet(ctx, `namespace = $1 AND id = $2 AND status = 'active'`, namespace, id)
}

func (q *Queries) DestroyFleet(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, "UPDATE fleets SET status = 'destroyed' WHERE id = $1", id)
	return err
}

func (q *Queries) getFleet(ctx context.Context, where string, args ...any) (*api.Fleet, error) {
	row := q.db.QueryRow(ctx, fmt.Sprintf(`SELECT id, namespace, name, created_at, status, metadata FROM fleets WHERE %s`, where), args...)
	return scanFleet(row)
}

// UpdateFleetMetadata updates the metadata for a fleet
func (q *Queries) UpdateFleetMetadata(ctx context.Context, fleetID string, metadata api.Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = q.db.Exec(ctx, `UPDATE fleets SET metadata = $1 WHERE id = $2`, metadataJSON, fleetID)
	return err
}
