package corrosion

import (
	"context"
	"errors"
	"time"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (c *CorrosionClusterState) UpsertInstance(ctx context.Context, i cluster.MachineInstance) error {
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{
		{
			Query: `INSERT INTO instances (id, node, machine_id, machine_version, status, created_at, updated_at, local_ipv4)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
					ON CONFLICT (id, machine_id) DO UPDATE SET
						status = $5,
						updated_at = $7,
						local_ipv4 = $8`,

			Params: []any{i.Id, i.Node, i.MachineId, i.MachineVersion, i.Status, i.CreatedAt.Unix(), i.UpdatedAt.Unix(), i.LocalIPV4},
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

func (c *CorrosionClusterState) GetInstance(ctx context.Context, id string) (*cluster.MachineInstance, error) {
	row, err := c.corroclient.QueryRow(ctx, corroclient.Statement{
		Query:  `SELECT id, node, machine_id, machine_version, status, created_at, updated_at, local_ipv4 FROM instances WHERE id = $1`,
		Params: []interface{}{id},
	})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return nil, errdefs.NewNotFound("instance not found")
		}
		return nil, err
	}

	var i cluster.MachineInstance
	var createdAt int64
	var updatedAt int64

	err = row.Scan(&i.Id, &i.Node, &i.MachineId, &i.MachineVersion, &i.Status, &createdAt, &updatedAt, &i.LocalIPV4)
	if err != nil {
		return nil, err
	}

	i.CreatedAt = time.Unix(createdAt, 0)
	i.UpdatedAt = time.Unix(updatedAt, 0)

	return &i, nil
}

func (c *CorrosionClusterState) WatchInstance(ctx context.Context, machineId string, instanceId string) (context.CancelFunc, <-chan cluster.MachineInstance, error) {
	sub, err := c.corroclient.PostSubscription(ctx, corroclient.Statement{
		Query:  `SELECT id, node, machine_id, machine_version, status, created_at, updated_at FROM instances WHERE machine_id = $1 AND id = $2`,
		Params: []any{machineId, instanceId},
	})
	if err != nil {
		return nil, nil, err
	}

	updates := make(chan cluster.MachineInstance)

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

					var i cluster.MachineInstance

					var createdAt int64
					var updatedAt int64

					err := row.Scan(&i.Id, &i.Node, &i.MachineId, &i.MachineVersion, &i.Status, &createdAt, &updatedAt)
					if err != nil {
						continue
					}

					i.CreatedAt = time.Unix(createdAt, 0)
					i.UpdatedAt = time.Unix(updatedAt, 0)
					updates <- i

				}
			}

		}
	}()

	return cancel, updates, nil
}
