package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/internal/dbutil"
	"github.com/valyentdev/ravel/ravel/state/db/schema"
)

func scanFleet(row dbutil.Scannable) (*api.Fleet, error) {
	var fleet api.Fleet
	err := row.Scan(&fleet.Id, &fleet.Namespace, &fleet.Name, &fleet.CreatedAt, &fleet.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errdefs.NewNotFound("fleet not found")
		}

		return nil, err
	}
	return &fleet, err
}

func (q *Queries) CreateFleet(ctx context.Context, fleet api.Fleet) error {
	_, err := q.db.Exec(ctx, `INSERT INTO fleets (id, namespace, name, created_at) VALUES ($1, $2, $3, $4)`, fleet.Id, fleet.Namespace, fleet.Name, fleet.CreatedAt)
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

	return err
}

func (q *Queries) ListFleets(ctx context.Context, namespace string) ([]api.Fleet, error) {
	rows, err := q.db.Query(ctx, `SELECT id, namespace, name, created_at, status FROM fleets WHERE namespace = $1 AND status = 'active'`, namespace)
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
	row := q.db.QueryRow(ctx, fmt.Sprintf(`SELECT id, namespace, name, created_at, status FROM fleets WHERE %s`, where), args...)
	return scanFleet(row)
}
