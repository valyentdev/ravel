package corrosion

import (
	"context"

	"github.com/valyentdev/ravel/api"
)

func (c *Queries) CreateGateway(ctx context.Context, gateway api.Gateway) error {
	_, err := c.dbtx.Exec(
		ctx,
		`INSERT INTO gateways (id, namespace, fleet_id, name, target_port) VALUES ($1, $2, $3, $4, $5)`,
		gateway.Id, gateway.Namespace, gateway.FleetId, gateway.Name, gateway.TargetPort,
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *Queries) DeleteFleetGateways(ctx context.Context, id string) error {
	_, err := c.dbtx.Exec(ctx, `DELETE FROM gateways WHERE fleet_id = $1`, id)
	if err != nil {
		return err
	}

	return nil
}

func (c *Queries) DeleteGateway(ctx context.Context, id string) error {
	_, err := c.dbtx.Exec(ctx, `DELETE FROM gateways WHERE id = $1`, id)
	if err != nil {
		return err
	}
	return nil
}
