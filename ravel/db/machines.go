package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/internal/dbutil"
)

func scanMachine(s dbutil.Scannable) (m cluster.Machine, err error) {
	var version string
	err = s.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &version, &m.Region, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = errdefs.NewNotFound("machine not found")
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

func (q *Queries) CreateMachine(ctx context.Context, machine cluster.Machine) error {

	_, err := q.db.Exec(ctx, `INSERT INTO machines (id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, machine.Id, machine.Namespace, machine.FleetId, machine.Node, machine.InstanceId, machine.MachineVersion.String(), machine.Region, machine.CreatedAt, machine.UpdatedAt)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerr.ConstraintName == "machines_pkey" {
				return errdefs.NewAlreadyExists("machine already exists")
			}
			if pgerr.ConstraintName == "machines_fleet_id_fkey" {
				return errdefs.NewNotFound("fleet not found")
			}
		}
		return err
	}
	return nil
}

const updateMachineQuery = `
UPDATE machines SET
	node = $1,
	instance_id = $2,
	machine_version = $3,
	updated_at = $4 
WHERE id = $5`

func (q *Queries) UpdateMachine(ctx context.Context, machine cluster.Machine) error {
	_, err := q.db.Exec(
		ctx,
		updateMachineQuery,
		machine.Node,
		machine.InstanceId,
		machine.MachineVersion.String(),
		machine.UpdatedAt,
		machine.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

const baseSelectMachine = `SELECT id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at FROM machines`

func (q *Queries) ListMachines(ctx context.Context, fleetId string) ([]cluster.Machine, error) {
	rows, err := q.db.Query(ctx, fmt.Sprintf("%s WHERE fleet_id = $1", baseSelectMachine), fleetId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	machines := []cluster.Machine{}
	for rows.Next() {
		machine, err := scanMachine(rows)
		if err != nil {
			return nil, err
		}
		machines = append(machines, machine)
	}

	return machines, nil
}

func (q *Queries) GetMachine(ctx context.Context, namespace, fleetId, id string) (cluster.Machine, error) {
	row := q.db.QueryRow(ctx, fmt.Sprintf("%s WHERE namespace = $1 AND fleet_id = $2 AND id = $3", baseSelectMachine), namespace, fleetId, id)
	machine, err := scanMachine(row)
	if err != nil {
		return machine, err
	}

	return machine, nil
}

func (q *Queries) DestroyMachine(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, `UPDATE machines SET destroyed = TRUE, updated_at = $1 WHERE  id = $2`, time.Now(), id)
	if err != nil {
		return err
	}

	return nil
}

const insertEventQuery = `INSERT INTO machine_events (id, type, origin, payload, instance_id, machine_id, status, timestamp) 
								VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
								ON CONFLICT DO NOTHING`

func (q *Queries) PushMachineEvent(ctx context.Context, event api.MachineEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	_, err = q.db.Exec(ctx, insertEventQuery, event.Id.String(), event.Type, event.Origin, json.RawMessage(payload), event.InstanceId, event.MachineId, event.Status, event.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

func (q *Queries) ListMachineEvents(ctx context.Context, machineId string) ([]api.MachineEvent, error) {
	rows, err := q.db.Query(ctx, `SELECT id, type, origin, payload, instance_id, machine_id, status, timestamp FROM machine_events WHERE machine_id = $1`, machineId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []api.MachineEvent{}
	for rows.Next() {
		var payload json.RawMessage
		var event api.MachineEvent
		err := rows.Scan(&event.Id, &event.Type, &event.Origin, &payload, &event.InstanceId, &event.MachineId, &event.Status, &event.Timestamp)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(payload, &event.Payload)
		if err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	return events, nil

}
