package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/pkg/core"
)

func scanMachine(s scannable) (m core.Machine, err error) {
	var version string
	err = s.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &version, &m.Region, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = core.NewNotFound("machine not found")
		}
		return
	}

	machineVersion, err := ulid.Parse(version)
	if err != nil {
		return
	}

	m.MachineVersion = machineVersion

	return
}

func (q *Queries) CreateMachine(ctx context.Context, machine core.Machine) error {

	_, err := q.db.Exec(ctx, `INSERT INTO machines (id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, machine.Id, machine.Namespace, machine.FleetId, machine.Node, machine.InstanceId, machine.MachineVersion.String(), machine.Region, machine.CreatedAt, machine.UpdatedAt)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "machines_pkey" {
				return core.NewAlreadyExists("machine already exists")
			}
			if pgerr.ConstraintName == "machines_fleet_id_fkey" {
				return core.NewNotFound("fleet not found")
			}
		}
		return err
	}
	return nil
}

func (q *Queries) UpdateMachine(ctx context.Context, machine core.Machine) error {
	return nil
}

const baseSelectMachine = `SELECT id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at FROM machines`

func (q *Queries) ListMachines(ctx context.Context, fleetId string) ([]core.Machine, error) {
	rows, err := q.db.Query(ctx, fmt.Sprintf("%s WHERE fleet_id = $1", baseSelectMachine), fleetId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []core.Machine{}
	for rows.Next() {
		machine, err := scanMachine(rows)
		if err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	return machines, nil
}

func (q *Queries) GetMachine(ctx context.Context, namespace, fleetId, id string) (core.Machine, error) {
	row := q.db.QueryRow(ctx, fmt.Sprintf("%s WHERE namespace = $1 AND fleet_id = $2 AND id = $3", baseSelectMachine), namespace, fleetId, id)
	machine, err := scanMachine(row)
	if err != nil {
		return machine, err
	}

	return machine, nil
}

func (q *Queries) DestroyMachine(ctx context.Context, namespace, fleetId, id string) error {
	_, err := q.db.Exec(ctx, `UPDATE machines SET destroyed = TRUE WHERE namespace = $1 AND fleet_id = $2 AND id = $3 `, namespace, fleetId, id)
	if err != nil {
		return err
	}

	return nil
}
