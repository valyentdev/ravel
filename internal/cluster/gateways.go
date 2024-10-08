package cluster

import (
	"context"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/pkg/core"
)

func (c *ClusterState) CreateGateway(ctx context.Context, gateway core.Gateway) error {
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{{
		Query:  `INSERT INTO gateways (id, namespace, fleet_id, name, target_port) VALUES ($1, $2, $3, $4, $5)`,
		Params: []any{gateway.Id, gateway.Namespace, gateway.FleetId, gateway.Name, gateway.TargetPort},
	}})
	if err != nil {
		return err
	}

	if result.Errors() != nil {
		return result.Errors()[0]
	}

	return nil
}

func (c *ClusterState) DeleteGateway(ctx context.Context, id string) error {
	result, err := c.corroclient.Exec(ctx, []corroclient.Statement{{
		Query:  `DELETE FROM gateways WHERE id = $1`,
		Params: []any{id},
	}})
	if err != nil {
		return err
	}

	if result.Errors() != nil {
		return result.Errors()[0]
	}

	return nil
}
