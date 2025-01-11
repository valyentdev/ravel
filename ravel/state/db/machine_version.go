package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func buildInsertMVQuery(mv *api.MachineVersion) (string, []interface{}, error) {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return "", nil, err
	}

	resourcesBytes, err := json.Marshal(mv.Resources)
	if err != nil {
		return "", nil, err
	}

	query := `INSERT INTO machine_versions (id, machine_id, config, resources, namespace) VALUES ($1, $2, $3, $4, $5)`
	args := []any{mv.Id, mv.MachineId, configBytes, resourcesBytes, mv.Namespace}
	return query, args, nil
}

func (q *Queries) CreateMachineVersion(ctx context.Context, mv api.MachineVersion) error {
	query, args, err := buildInsertMVQuery(&mv)
	if err != nil {
		return err
	}
	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "machine_versions_machine_id_fkey" {
				return errdefs.NewNotFound("machine not found")
			}
		}
		return err
	}
	return nil
}

func (q *Queries) ListMachineVersions(ctx context.Context, machineId string) ([]api.MachineVersion, error) {
	rows, err := q.db.Query(ctx, `SELECT id, machine_id, config, resources, namespace FROM machine_versions WHERE machine_id = $1 ORDER BY id DESC`, machineId)
	if err != nil {
		return nil, err
	}

	mvs := []api.MachineVersion{}

	for rows.Next() {
		var mv api.MachineVersion
		var configBytes []byte
		var resourcesBytes []byte
		err := rows.Scan(&mv.Id, &mv.MachineId, &configBytes, &resourcesBytes, &mv.Namespace)
		if err != nil {
			if err == pgx.ErrNoRows {
				return mvs, nil
			}
			return nil, err
		}

		err = json.Unmarshal(configBytes, &mv.Config)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(resourcesBytes, &mv.Resources)
		if err != nil {
			return nil, err
		}

		mvs = append(mvs, mv)
	}

	return mvs, nil
}
