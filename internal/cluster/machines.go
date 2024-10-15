package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/oklog/ulid"
	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/internal/dbutil"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api"
)

func (c *ClusterState) CreateMachine(ctx context.Context, m core.Machine, mv core.MachineVersion) error {
	configBytes, err := json.Marshal(mv.Config)
	if err != nil {
		return err
	}
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{
		{
			Query: `INSERT INTO machines 
					(id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, machine_version)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			Params: []any{m.Id, m.Namespace, m.FleetId, m.Node, m.InstanceId, m.Region, m.CreatedAt.Unix(), m.UpdatedAt.Unix(), m.MachineVersion.String()},
		},
		{
			Query: `INSERT INTO machine_versions
					(id, machine_id, config)
					VALUES ($1, $2, $3)`,
			Params: []any{mv.Id, mv.MachineId, string(configBytes)},
		},
	})

	if err != nil {
		return err
	}

	errs := result.Errors()

	err = errors.Join(errs...)

	return err
}

func (c *ClusterState) GetMachine(ctx context.Context, namespace string, fleetId string, id string, destroyed bool) (*core.Machine, error) {
	row, err := c.corroclient.QueryRow(ctx,
		corroclient.Statement{
			Query: `SELECT id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, destroyed
					FROM machines
					WHERE namespace = $1 AND fleet_id = $2 AND id = $3 AND destroyed = $4`,
			Params: []any{namespace, fleetId, id, 0},
		})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return nil, core.NewNotFound("machine not found")
		}
		return nil, err
	}

	var m core.Machine
	var createdAt int64
	var updatedAt int64

	err = row.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &m.Region, &createdAt, &updatedAt, &m.Destroyed)
	if err != nil {
		return nil, err
	}

	m.CreatedAt = time.Unix(createdAt, 0)
	m.UpdatedAt = time.Unix(updatedAt, 0)

	return &m, nil
}

func (c *ClusterState) ListMachines(ctx context.Context, namespace string, fleet string, destroyed bool) ([]core.Machine, error) {
	rows, err := c.corroclient.Query(ctx,
		corroclient.Statement{
			Query: `SELECT id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, destroyed
					FROM machines
					WHERE namespace = $1 AND fleet_id = $2 AND destroyed = $3`,
			Params: []any{namespace, fleet, destroyed},
		})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []core.Machine{}, nil
		}
		return nil, err
	}

	var machines []core.Machine
	for rows.Next() {
		var m core.Machine
		var createdAt int64
		var updatedAt int64

		err := rows.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &m.Region, &createdAt, &updatedAt, &m.Destroyed)
		if err != nil {
			return nil, err
		}

		m.CreatedAt = time.Unix(createdAt, 0)
		m.UpdatedAt = time.Unix(updatedAt, 0)
		machines = append(machines, m)
	}
	return machines, nil
}

func (c *ClusterState) DestroyMachine(ctx context.Context, id string) error {
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{
		{
			Query: `UPDATE machines
					SET destroyed = true
					SET updated_at = $1
					WHERE id = $2`,
			Params: []any{time.Now().UTC().Unix(), id},
		},
	})

	if err != nil {
		return err
	}

	if err := result.Results[0].Err(); err != nil {
		return err
	}

	return err
}

const baseSelectAPIMachine = `SELECT m.id, m.namespace, m.fleet_id, m.instance_id, m.machine_version, m.region, m.created_at, m.updated_at, i.status, mv.config
							  FROM machines m
							  JOIN instances i ON m.instance_id = i.id
							  JOIN machine_versions mv ON m.machine_version = mv.id`

func scanAPIMachine(row dbutil.Scannable) (*api.Machine, error) {
	var m api.Machine
	var config []byte
	var createdAt int64
	var updatedAt int64

	var state string
	var version string

	err := row.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.InstanceId, &version, &m.Region, &createdAt, &updatedAt, &state, &config)
	if err != nil {
		return nil, err
	}

	m.State = core.MachineStatus(state)

	err = json.Unmarshal(config, &m.Config)
	if err != nil {
		return nil, err
	}

	machineVersion, err := ulid.Parse(version)
	if err != nil {
		return nil, err
	}

	m.MachineVersion = machineVersion

	m.CreatedAt = time.Unix(createdAt, 0)
	m.UpdatedAt = time.Unix(updatedAt, 0)

	return &m, nil
}

func (c *ClusterState) GetAPIMachine(ctx context.Context, namespace, fleetId, machineId string) (*api.Machine, error) {
	row, err := c.corroclient.QueryRow(
		ctx,
		corroclient.Statement{
			Query: baseSelectAPIMachine + `
			WHERE m.namespace = $1 AND m.fleet_id = $2 AND m.id = $3`,
			Params: []any{namespace, fleetId, machineId},
		},
	)
	if err != nil {
		if err == corroclient.ErrNoRows {
			return nil, core.NewNotFound("machine not found")
		}
	}

	return scanAPIMachine(row)
}

func (c *ClusterState) ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error) {
	query := baseSelectAPIMachine + `
			WHERE m.namespace = $1 AND m.fleet_id = $2`
	if !includeDestroyed {
		query += ` AND i.status != 'destroyed'`
	}

	rows, err := c.corroclient.Query(
		ctx,
		corroclient.Statement{
			Query:  query,
			Params: []any{namespace, fleetId},
		},
	)
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []api.Machine{}, nil
		}
		return nil, err
	}

	var machines []api.Machine
	for rows.Next() {
		m, err := scanAPIMachine(rows)
		if err != nil {
			return nil, err
		}
		machines = append(machines, *m)
	}
	return machines, nil
}
