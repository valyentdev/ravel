package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/corroclient"
)

type Instance struct {
	Id        string `json:"id"`
	MachineId string `json:"machine_id"`
	Node      string `json:"node"`
	Config    core.MachineConfig
	ImageRef  string    `json:"image_ref"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InstanceStatus struct {
	Id        string             `json:"id"`
	MachineId string             `json:"machine_id"`
	Status    core.MachineStatus `json:"status"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func (c *ClusterState) UpsertInstance(ctx context.Context, i Instance, status InstanceStatus) error {
	configBytes, err := json.Marshal(i.Config)
	if err != nil {
		return err
	}

	result, err := c.corroclient.ExecContext(ctx, []corroclient.Statement{
		{
			Query: `INSERT INTO instances
					(id, machine_id, node, config, image_ref, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $6)
					ON CONFLICT (id, machine_id)
					DO UPDATE SET updated_at = $6, config =$4`,
			Params: []any{i.Id, i.MachineId, i.Node, string(configBytes), i.ImageRef, i.UpdatedAt.Unix()},
		},
		{
			Query: `INSERT INTO instance_statuses 
					(id, machine_id, status, updated_at)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (id, machine_id)
					DO UPDATE SET status = $3, updated_at = $4
					`,
			Params: []any{status.Id, status.MachineId, status.Status, status.UpdatedAt.Unix()},
		},
	})
	if err != nil {
		return err
	}

	var errs []error

	for _, r := range result.Results {
		if err := r.Err(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

type InstanceWithStatus struct {
	Instance Instance
	Status   InstanceStatus
}

func (c ClusterState) ListInstances(ctx context.Context, machineId string) ([]InstanceWithStatus, error) {
	rows, err := c.corroclient.QueryContext(ctx, corroclient.Statement{
		Query: `SELECT i.id, i.machine_id, i.node, i.config, i.image_ref, i.created_at, i.updated_at, s.status, s.updated_at as status_updated_at, s.id as status_id
				FROM instances i
				JOIN instance_statuses s ON i.id = s.id
				WHERE i.machine_id = $1`,
		Params: []any{machineId},
	})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []InstanceWithStatus{}, nil
		}
		return nil, err
	}

	var instances []InstanceWithStatus
	for rows.Next() {
		var i Instance
		var s InstanceStatus
		var statusUpdatedAt int64
		var createdAt int64
		var updatedAt int64
		var configBytes []byte

		err := rows.Scan(&i.Id, &i.MachineId, &i.Node, &configBytes, &i.ImageRef, &createdAt, &updatedAt, &s.Status, &statusUpdatedAt, &s.Id)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configBytes, &i.Config); err != nil {
			return nil, err
		}

		i.CreatedAt = time.Unix(createdAt, 0)
		i.UpdatedAt = time.Unix(updatedAt, 0)
		instances = append(instances, InstanceWithStatus{
			Instance: i,
			Status: InstanceStatus{
				Id:        s.Id,
				MachineId: i.MachineId,
				Status:    s.Status,
				UpdatedAt: time.Unix(statusUpdatedAt, 0),
			},
		})
	}

	return instances, nil
}

func (c *ClusterState) GetCurrentMachineInstance(ctx context.Context, machineId string) (*InstanceWithStatus, error) {
	row, err := c.corroclient.QueryRowContext(ctx, corroclient.Statement{
		Query: `SELECT i.id, i.machine_id, i.node, i.config, i.image_ref, i.created_at, i.updated_at, s.status, s.updated_at as status_updated_at, s.id as status_id
				FROM instances i
				JOIN instance_statuses s ON i.id = s.id
				WHERE i.machine_id = $1
				ORDER BY i.updated_at DESC
				LIMIT 1`,
		Params: []interface{}{machineId},
	})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var i Instance
	var s InstanceStatus
	var status string
	var statusUpdatedAt int64
	var createdAt int64
	var updatedAt int64
	var configBytes []byte

	err = row.Scan(&i.Id, &i.MachineId, &i.Node, &configBytes, &i.ImageRef, &createdAt, &updatedAt, &status, &statusUpdatedAt, &s.Id)
	if err != nil {
		return nil, err
	}

	s.Status = core.MachineStatus(status)

	if err := json.Unmarshal(configBytes, &i.Config); err != nil {
		return nil, err
	}

	i.CreatedAt = time.Unix(createdAt, 0)
	i.UpdatedAt = time.Unix(updatedAt, 0)

	return &InstanceWithStatus{
		Instance: i,
		Status: InstanceStatus{
			Id:        s.Id,
			MachineId: i.MachineId,
			Status:    s.Status,
			UpdatedAt: time.Unix(statusUpdatedAt, 0),
		},
	}, nil
}

func (c *ClusterState) WatchInstance(ctx context.Context, machineId string, instanceId string) (context.CancelFunc, <-chan InstanceWithStatus, error) {
	sub, err := c.corroclient.PostSubscription(ctx, corroclient.Statement{
		Query: `SELECT i.id, i.machine_id, i.node, i.config, i.image_ref, i.created_at, i.updated_at, s.status, s.updated_at as status_updated_at, s.id as status_id
				FROM instances i
				JOIN instance_statuses s ON i.id = s.id
				WHERE i.machine_id = $1 AND i.id = $2`,
		Params: []any{machineId, instanceId},
	})
	if err != nil {
		return nil, nil, err
	}

	updates := make(chan InstanceWithStatus)

	subCtx, cancel := context.WithCancel(context.Background())

	go func() {
		events := sub.Events()
		for {
			select {
			case <-subCtx.Done():
				sub.Close()
				return
			case e := <-events:
				if e.Type() == corroclient.EventTypeError {
					sub.Close()
					return
				}

				if e.Type() == corroclient.EventTypeRow || e.Type() == corroclient.EventTypeChange {
					var row *corroclient.Row
					if e.Type() == corroclient.EventTypeRow {
						row = e.(*corroclient.Row)
					} else {
						change := e.(*corroclient.Change)
						row = change.Row

					}

					var i InstanceWithStatus

					var statusUpdatedAt int64
					var createdAt int64
					var updatedAt int64
					var configBytes []byte

					err := row.Scan(&i.Instance.Id, &i.Instance.MachineId, &i.Instance.Node, &configBytes, &i.Instance.ImageRef, &createdAt, &updatedAt, &i.Status.Status, &statusUpdatedAt, &i.Status.Id)
					if err != nil {
						continue
					}

					if err := json.Unmarshal(configBytes, &i.Instance.Config); err != nil {
						continue
					}

					i.Instance.CreatedAt = time.Unix(createdAt, 0)
					i.Instance.UpdatedAt = time.Unix(updatedAt, 0)
					i.Status.UpdatedAt = time.Unix(statusUpdatedAt, 0)

					updates <- i

				}
			}

		}
	}()

	return cancel, updates, nil
}
