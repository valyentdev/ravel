package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/valyentdev/ravel/pkg/core"
)

var instanceFields = []string{
	"id",
	"namespace",
	"fleet_id",
	"node_id",
	"machine_id",
	"reservation_id",
	"desired_status",
	"restarts",
	"config",
	"created_at",
	"prepared",
	"destroyed",
	"destroyed_at",
}

var instanceColumns = allColumns(instanceFields)
var instancePlaceholders = placeholders(len(instanceFields))

func scanInstance(row scannable) (*core.Instance, error) {
	var instance core.Instance

	var destroyedAt sql.NullTime

	configJSON := []byte{}
	err := row.Scan(
		&instance.Id,
		&instance.Namespace,
		&instance.FleetId,
		&instance.NodeId,
		&instance.MachineId,
		&instance.ReservationId,
		&instance.DesiredStatus,
		&instance.Restarts,
		&configJSON,
		&instance.CreatedAt,
		&instance.Prepared,
		&instance.Destroyed,
		&destroyedAt,
	)

	if destroyedAt.Valid {
		instance.DestroyedAt = destroyedAt.Time
	}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(configJSON, &instance.Config)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (q *Queries) CreateInstance(ctx context.Context, instance core.Instance) error {
	var createInstanceRequest = fmt.Sprintf("INSERT INTO instances (%s) VALUES (%s)", instanceColumns, instancePlaceholders)

	configJSON, err := json.Marshal(instance.Config)
	if err != nil {
		return err
	}

	_, err = q.db.ExecContext(ctx, createInstanceRequest,
		instance.Id,
		instance.Namespace,
		instance.FleetId,
		instance.NodeId,
		instance.MachineId,
		instance.ReservationId,
		instance.DesiredStatus,
		instance.Restarts,
		configJSON,
		instance.CreatedAt,
		instance.Prepared,
		instance.Destroyed,
		instance.DestroyedAt,
	)
	return err
}

func (q *Queries) GetInstance(ctx context.Context, id string) (*core.Instance, error) {
	request := fmt.Sprintf("SELECT %s FROM instances WHERE id = ?", instanceColumns)
	row := q.db.QueryRowContext(ctx, request, id)
	if row.Err() != nil {
		return nil, row.Err()
	}

	return scanInstance(row)
}

func (q *Queries) ListInstances(ctx context.Context) ([]core.Instance, error) {
	request := fmt.Sprintf("SELECT %s FROM instances WHERE destroyed = false", instanceColumns)
	rows, err := q.db.QueryContext(ctx, request)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}
	defer rows.Close()

	instances := []core.Instance{}
	for rows.Next() {
		instance, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, *instance)
	}

	return instances, nil
}

func (s *Queries) MarkInstanceDestroyed(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE instances SET destroyed = true WHERE id = ?1", id)
	return err
}

func (s *Queries) MarkInstanceAsPrepared(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE instances SET prepared = true WHERE id = ?1", id)
	return err
}

func (s *Queries) DestroyInstance(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM instances WHERE id = ?1", id)
	return err
}

func (s *Queries) IncrementInstanceRestarts(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE instances SET restarts = restarts + 1 WHERE id = ?1", id)
	return err
}

func (s *Queries) ResetRestarts(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE instances SET restarts = 0 WHERE id = ?1", id)
	return err
}

func (s *Queries) UpdateInstanceDesiredStatus(ctx context.Context, id string, status core.InstanceStatus) error {
	_, err := s.db.ExecContext(ctx, "UPDATE instances SET desired_status = ?1 WHERE id = ?2", status, id)
	return err
}
