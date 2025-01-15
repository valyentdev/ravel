package corrosion

import (
	"context"
	"encoding/json"

	"github.com/valyentdev/ravel/core/cluster"
)

type corroBool uint8

func (b corroBool) Bool() bool {
	return b == 1
}

func fromBool(b bool) corroBool {
	if b {
		return 1
	}
	return 0
}

func (c *Queries) UpsertInstance(ctx context.Context, i cluster.MachineInstance) error {
	eventsBytes, err := json.Marshal(i.Events)
	if err != nil {
		return err
	}
	_, err = c.dbtx.Exec(ctx,
		`INSERT INTO instances (id, node, machine_id, machine_version, status, created_at, updated_at, local_ipv4, events, namespace, enable_machine_gateway)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
					ON CONFLICT (id, machine_id) DO UPDATE SET
						status = $5,
						updated_at = $7,
						local_ipv4 = $8,
						events = $9,
						enable_machine_gateway = $11
						`,

		i.Id, i.Node, i.MachineId, i.MachineVersion, i.Status, i.CreatedAt.Unix(), i.UpdatedAt.Unix(), i.LocalIPV4, string(eventsBytes), i.Namespace, fromBool(i.EnableMachineGateway),
	)
	if err != nil {
		return err
	}

	return nil
}
