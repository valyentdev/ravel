package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (q *Queries) CreateMachineVersion(ctx context.Context, mv api.MachineVersion) error {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return err
	}

	resourcesBytes, err := json.Marshal(mv.Resources)
	if err != nil {
		return err
	}

	_, err = q.db.Exec(ctx, `INSERT INTO machine_versions (id, machine_id, config, resources) VALUES ($1, $2, $3, $4)`, mv.Id.String(), mv.MachineId, configBytes, resourcesBytes)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "machine_versions_pkey" {
				return errdefs.NewAlreadyExists("machine version already exists")
			}
			if pgerr.ConstraintName == "machine_versions_machine_id_fkey" {
				return errdefs.NewNotFound("machine not found")
			}
		}
		return err
	}
	return nil
}

func (q *Queries) ListMachineVersions(ctx context.Context, machineId string) ([]api.MachineVersion, error) {
	rows, err := q.db.Query(ctx, `SELECT id, machine_id, config, resources FROM machine_versions WHERE machine_id = $1 ORDER BY id DESC`, machineId)
	if err != nil {
		return nil, err
	}

	mvs := []api.MachineVersion{}

	for rows.Next() {
		var id string
		var mv api.MachineVersion
		var configBytes []byte
		var resourcesBytes []byte
		err := rows.Scan(&id, &mv.MachineId, &configBytes, &resourcesBytes)
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

		uid, err := ulid.Parse(id)
		if err != nil {
			return nil, err
		}

		mv.Id = uid

		mvs = append(mvs, mv)
	}

	return mvs, nil
}
