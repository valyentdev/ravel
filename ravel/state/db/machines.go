package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/internal/dbutil"
	"github.com/jackc/pgx/v5"
)

func scanMachine(s dbutil.Scannable) (m cluster.Machine, err error) {
	var metadataJSON []byte
	err = s.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &m.MachineVersion, &m.Region, &m.CreatedAt, &m.UpdatedAt, &m.DestroyedAt, &metadataJSON)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = errdefs.NewNotFound("machine not found")
		}
		return
	}

	// Deserialize metadata if present
	if len(metadataJSON) > 0 && string(metadataJSON) != "{}" {
		var metadata api.Metadata
		if json.Unmarshal(metadataJSON, &metadata) == nil {
			// Only set metadata if it has labels or annotations
			if len(metadata.Labels) > 0 || len(metadata.Annotations) > 0 {
				m.Metadata = &metadata
			}
		}
	}

	return
}

func (q *Queries) CreateMachine(ctx context.Context, machine cluster.Machine, mv api.MachineVersion) error {
	mvQueries, mvArgs, err := buildInsertMVQuery(&mv)
	if err != nil {
		return err
	}

	var metadataJSON []byte
	if machine.Metadata != nil {
		metadataJSON, err = json.Marshal(machine.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	} else {
		metadataJSON = []byte("{}")
	}

	queuedQueries := []*pgx.QueuedQuery{
		{
			SQL:       `INSERT INTO machines (id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			Arguments: []interface{}{machine.Id, machine.Namespace, machine.FleetId, machine.Node, machine.InstanceId, machine.MachineVersion, machine.Region, machine.CreatedAt, machine.UpdatedAt, metadataJSON},
		},
		{
			SQL:       mvQueries,
			Arguments: mvArgs,
		},
	}

	result := q.db.SendBatch(ctx, &pgx.Batch{
		QueuedQueries: queuedQueries,
	})
	if err := result.Close(); err != nil {
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
		machine.MachineVersion,
		machine.UpdatedAt,
		machine.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) CountExistingMachinesInFleet(ctx context.Context, fleetId string) (int, error) {
	var count int
	err := q.db.QueryRow(ctx, `SELECT COUNT(*) FROM machines WHERE fleet_id = $1 AND destroyed_at IS NULL`, fleetId).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

const baseSelectMachine = `SELECT id, namespace, fleet_id, node, instance_id, machine_version, region, created_at, updated_at, destroyed_at, metadata FROM machines`

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

func (q *Queries) GetMachine(ctx context.Context, namespace, fleetId, id string, showDestroyed bool) (cluster.Machine, error) {
	where := fmt.Sprintf("%s WHERE namespace = $1 AND fleet_id = $2 AND id = $3", baseSelectMachine)
	if !showDestroyed {
		where += " AND destroyed_at IS NULL"
	}

	row := q.db.QueryRow(ctx, where, namespace, fleetId, id)
	machine, err := scanMachine(row)
	if err != nil {
		return machine, err
	}

	return machine, nil
}

func (q *Queries) DestroyMachine(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, `UPDATE machines SET destroyed_at = $1, updated_at = $1 WHERE  id = $2`, time.Now(), id)
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

	_, err = q.db.Exec(ctx, insertEventQuery, event.Id, event.Type, event.Origin, json.RawMessage(payload), event.InstanceId, event.MachineId, event.Status, event.Timestamp)
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

// UpdateMachineMetadata updates the metadata for a machine
func (q *Queries) UpdateMachineMetadata(ctx context.Context, machineID string, metadata api.Metadata) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = q.db.Exec(ctx, `UPDATE machines SET metadata = $1, updated_at = $2 WHERE id = $3`, metadataJSON, time.Now(), machineID)
	return err
}
