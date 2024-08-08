package cluster

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/corroclient"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

type Machine struct {
	Id         string    `json:"id"`
	Namespace  string    `json:"namespace"`
	FleetId    string    `json:"fleet_id"`
	Node       string    `json:"node"`
	InstanceId string    `json:"instance_id"`
	Region     string    `json:"region"`
	CreatedAd  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Destroyed  bool      `json:"destroyed"`
}

func (c *ClusterState) CreateMachine(ctx context.Context, m Machine) error {
	result, err := c.corroclient.ExecContext(ctx, []corroclient.Statement{
		{
			Query: `INSERT INTO machines 
					(id, namespace, fleet_id, node, instance_id, region, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
			Params: []any{m.Id, m.Namespace, m.FleetId, m.Node, m.InstanceId, m.Region, time.Now().UTC().Unix()},
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

func (c *ClusterState) GetMachine(ctx context.Context, namespace string, fleet string, id string, destroyed bool) (*Machine, error) {
	row, err := c.corroclient.QueryRowContext(ctx,
		corroclient.Statement{
			Query: `SELECT id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, destroyed
					FROM machines
					WHERE namespace = $1 AND fleet_id = $2 AND id = $3 AND destroyed = $4`,
			Params: []any{namespace, fleet, id, destroyed},
		})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return nil, ravelerrors.NewNotFound("machine not found")
		}
		return nil, err
	}

	var m Machine
	var createdAt int64
	var updatedAt int64

	err = row.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &m.Region, &createdAt, &updatedAt, &m.Destroyed)
	if err != nil {
		return nil, err
	}

	m.CreatedAd = time.Unix(createdAt, 0)
	m.UpdatedAt = time.Unix(updatedAt, 0)

	return &m, nil
}

func (c *ClusterState) ListMachines(ctx context.Context, namespace string, fleet string, destroyed bool) ([]Machine, error) {
	rows, err := c.corroclient.QueryContext(ctx,
		corroclient.Statement{
			Query: `SELECT id, namespace, fleet_id, node, instance_id, region, created_at, updated_at, destroyed
					FROM machines
					WHERE namespace = $1 AND fleet_id = $2 AND destroyed = $3`,
			Params: []any{namespace, fleet, destroyed},
		})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []Machine{}, nil
		}
		return nil, err
	}

	var machines []Machine
	for rows.Next() {
		var m Machine
		var createdAt int64
		var updatedAt int64

		err := rows.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.Node, &m.InstanceId, &m.Region, &createdAt, &updatedAt, &m.Destroyed)
		if err != nil {
			return nil, err
		}

		m.CreatedAd = time.Unix(createdAt, 0)
		m.UpdatedAt = time.Unix(updatedAt, 0)
		machines = append(machines, m)
	}
	return machines, nil
}

func (c *ClusterState) DestroyMachine(ctx context.Context, id string) error {
	result, err := c.corroclient.ExecContext(ctx, []corroclient.Statement{
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

const baseSelectAPIMachine = `
SELECT m.id, m.namespace, m.fleet_id, m.instance_id, m.region, i.config, i.updated_at, m.created_at, s.status
FROM machines m
JOIN instances i ON m.instance_id = i.id
JOIN instance_statuses s ON i.id = s.id
`

func (c *ClusterState) GetAPIMachine(ctx context.Context, namespace string, fleetId string, machineId string) (core.Machine, error) {
	var m core.Machine
	var config []byte
	var createdAt int64
	var updatedAt int64

	row, err := c.corroclient.QueryRowContext(
		ctx,
		corroclient.Statement{
			Query: baseSelectAPIMachine + `
			WHERE m.namespace = $1 AND m.fleet_id = $2 AND m.id = $3`,
			Params: []any{namespace, fleetId, machineId},
		},
	)
	if err != nil {
		if err == corroclient.ErrNoRows {
			return core.Machine{}, ravelerrors.NewNotFound("machine not found")
		}
	}

	var state string

	err = row.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.InstanceId, &m.Region, &config, &updatedAt, &createdAt, &state)
	if err != nil {
		return core.Machine{}, err
	}

	m.State = core.MachineStatus(state)

	err = json.Unmarshal(config, &m.Config)
	if err != nil {
		return core.Machine{}, err
	}

	m.CreatedAt = time.Unix(createdAt, 0)
	m.UpdatedAt = time.Unix(updatedAt, 0)

	return m, nil
}

func (c *ClusterState) ListAPIMachines(ctx context.Context, namespace string, fleetId string) ([]core.Machine, error) {
	rows, err := c.corroclient.QueryContext(
		ctx,
		corroclient.Statement{
			Query: baseSelectAPIMachine + `
			WHERE m.namespace = $1 AND m.fleet_id = $2`,
			Params: []any{namespace, fleetId},
		},
	)
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []core.Machine{}, nil
		}
		return nil, err
	}

	var machines []core.Machine
	for rows.Next() {
		var m core.Machine
		var config []byte

		var createdAt int64
		var updatedAt int64
		var state string

		err := rows.Scan(&m.Id, &m.Namespace, &m.FleetId, &m.InstanceId, &m.Region, &config, &createdAt, &updatedAt, &state)
		if err != nil {
			return nil, err
		}

		m.State = core.MachineStatus(state)

		err = json.Unmarshal(config, &m.Config)
		if err != nil {
			slog.Error("failed to unmarshal config", "err", err)

			return nil, err
		}

		m.CreatedAt = time.Unix(createdAt, 0)
		m.UpdatedAt = time.Unix(updatedAt, 0)

		machines = append(machines, m)
	}
	return machines, nil
}
