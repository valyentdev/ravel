package corrosion

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/instance"
)

func (c *CorrosionClusterState) UpsertInstance(ctx context.Context, i cluster.MachineInstance) error {
	eventsBytes, err := json.Marshal(i.Events)
	if err != nil {
		return err
	}
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{
		{
			Query: `INSERT INTO instances (id, node, machine_id, machine_version, status, created_at, updated_at, local_ipv4, events, namespace)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
					ON CONFLICT (id, machine_id) DO UPDATE SET
						status = $5,
						updated_at = $7,
						local_ipv4 = $8,
						events = $9
						`,

			Params: []any{i.Id, i.Node, i.MachineId, i.MachineVersion, i.Status, i.CreatedAt.Unix(), i.UpdatedAt.Unix(), i.LocalIPV4, string(eventsBytes), i.Namespace},
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

func (c *CorrosionClusterState) WatchInstanceStatus(ctx context.Context, machineId string, instanceId string) (<-chan instance.InstanceStatus, error) {
	sub, err := c.corroclient.PostSubscription(ctx, corroclient.Statement{
		Query:  `SELECT status FROM instances WHERE machine_id = $1 AND id = $2`,
		Params: []any{machineId, instanceId},
	})
	if err != nil {
		return nil, err
	}

	updates := make(chan instance.InstanceStatus)

	go func() {
		events := sub.Events()
		for {
			select {
			case <-ctx.Done():
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

					var status string

					err := row.Scan(&status)
					if err != nil {
						continue
					}

					updates <- instance.InstanceStatus(status)

				}
			}

		}
	}()

	return updates, nil
}
