package db

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/pkg/core"
)

func (q *Queries) CreateMachineVersion(ctx context.Context, mv core.MachineVersion) error {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return err
	}

	_, err = q.db.Exec(ctx, `INSERT INTO machine_versions (id, machine_id, config) VALUES ($1, $2, $3)`, mv.Id.String(), mv.MachineId, configBytes)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "machine_versions_pkey" {
				return core.NewAlreadyExists("machine version already exists")
			}
			if pgerr.ConstraintName == "machine_versions_machine_id_fkey" {
				return core.NewNotFound("machine not found")
			}
		}
		return err
	}
	return nil
}

func (q *Queries) ListMachineVersions(ctx context.Context, machineId string) ([]core.MachineVersion, error) {
	rows, err := q.db.Query(ctx, `SELECT id, machine_id, config FROM machine_versions WHERE machine_id = $1 ORDER BY id DESC`, machineId)
	if err != nil {
		return nil, err
	}

	mvs := []core.MachineVersion{}

	for rows.Next() {
		var id string
		var mv core.MachineVersion
		var configBytes []byte
		err := rows.Scan(&id, &mv.MachineId, &configBytes)
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

		uid, err := ulid.Parse(id)
		if err != nil {
			return nil, err
		}

		mv.Id = uid

		mvs = append(mvs, mv)
	}

	return mvs, nil
}
