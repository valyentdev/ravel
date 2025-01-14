package corrosion

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/internal/dbutil"
)

func (c *Queries) UpdateMachine(ctx context.Context, m cluster.Machine) error {
	_, err := c.dbtx.Exec(ctx,
		`UPDATE machines SET
				instance_id = $1,
				node = $2,
				machine_version = $3,
				updated_at = $4
				WHERE id = $5`,
		m.InstanceId, m.Node, m.MachineVersion, time.Now().UTC().Unix(), m.Id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *Queries) CreateMachineVersion(ctx context.Context, mv api.MachineVersion) error {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return err
	}

	resourcesBytes, err := json.Marshal(mv.Resources)
	if err != nil {
		return err
	}

	_, err = c.dbtx.Exec(ctx,
		`INSERT INTO machine_versions
				(id, machine_id, namespace, config, resources)
				VALUES ($1, $2, $3, $4, $5)`,
		mv.Id, mv.MachineId, mv.Namespace, configBytes, resourcesBytes,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Queries) CreateMachine(ctx context.Context, m cluster.Machine, mv api.MachineVersion) error {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return err
	}

	resourcesBytes, err := json.Marshal(mv.Resources)
	if err != nil {
		return err
	}

	queries := []*pgx.QueuedQuery{
		{
			SQL: `INSERT INTO machines
					(id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, machine_version)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			Arguments: []interface{}{m.Id, m.Namespace, m.FleetId, m.Node, m.InstanceId, m.Region, m.CreatedAt.Unix(), m.UpdatedAt.Unix(), m.MachineVersion},
		},
		{
			SQL: `INSERT INTO machine_versions
					(id, machine_id, namespace, config, resources)
					VALUES ($1, $2, $3, $4, $5)`,
			Arguments: []interface{}{mv.Id, mv.MachineId, mv.Namespace, configBytes, resourcesBytes},
		},
	}
	result := c.dbtx.SendBatch(ctx, &pgx.Batch{QueuedQueries: queries})
	if err := result.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Queries) DestroyMachine(ctx context.Context, id string) error {
	_, err := c.dbtx.Exec(ctx,
		`UPDATE machines SET
			destroyed_at = $1,
			updated_at = $1
			WHERE id = $2`,
		time.Now().UTC().Unix(), id,
	)
	if err != nil {
		return err
	}

	return nil
}

const baseSelectAPIMachine = `SELECT m.id, m.namespace, m.fleet_id, m.instance_id, m.machine_version, m.region, m.created_at, m.updated_at, i.status, mv.config, i.events, i.enable_machine_gateway
							  FROM machines m
							  JOIN instances i ON m.instance_id = i.id
							  JOIN machine_versions mv ON m.machine_version = mv.id`

func scanAPIMachine(row dbutil.Scannable) (*api.Machine, error) {
	var m api.Machine
	var config []byte
	var events []byte
	var createdAt int64
	var updatedAt int64
	var state string
	var gatewayEnabled corroBool
	err := row.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.InstanceId, &m.MachineVersion, &m.Region, &createdAt, &updatedAt, &state, &config, &events, &gatewayEnabled)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errdefs.NewNotFound("machine not found")
		}
		return nil, err
	}

	m.GatewayEnabled = gatewayEnabled.Bool()

	m.Status = api.MachineStatus(state)

	err = json.Unmarshal(config, &m.Config)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(events, &m.Events)
	if err != nil {
		return nil, err
	}

	m.CreatedAt = time.Unix(createdAt, 0)
	m.UpdatedAt = time.Unix(updatedAt, 0)

	return &m, nil
}

func (c *Queries) GetAPIMachine(ctx context.Context, namespace, fleetId, machineId string) (*api.Machine, error) {
	row := c.dbtx.QueryRow(
		ctx,
		baseSelectAPIMachine+`
			WHERE m.namespace = $1 AND m.fleet_id = $2 AND m.id = $3`,
		namespace, fleetId, machineId,
	)

	return scanAPIMachine(row)
}

func (c *Queries) ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error) {
	query := baseSelectAPIMachine + `
			WHERE m.namespace = $1 AND m.fleet_id = $2`
	if !includeDestroyed {
		query += ` AND i.status != 'destroyed'`
	}

	rows, err := c.dbtx.Query(ctx, query, namespace, fleetId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []api.Machine{}, nil
		}
		return nil, err
	}

	machines := []api.Machine{}
	for rows.Next() {
		m, err := scanAPIMachine(rows)
		if err != nil {
			return nil, err
		}
		machines = append(machines, *m)
	}
	return machines, nil
}
