package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

const baseSelectGateway = `SELECT id, name, namespace, fleet_id, protocol, target_port FROM gateways`

func (q Queries) DeleteFleetGateways(ctx context.Context, fleetId string) error {
	_, err := q.db.Exec(ctx, `DELETE FROM gateways WHERE fleet_id = $1`, fleetId)
	if err != nil {
		return err
	}
	return nil
}

func (q Queries) GetGatewayByName(ctx context.Context, namespace, name string) (gateway api.Gateway, err error) {
	err = q.db.QueryRow(ctx, baseSelectGateway+" WHERE namespace = $1 AND name = $2", namespace, name).Scan(
		&gateway.Id,
		&gateway.Name,
		&gateway.Namespace,
		&gateway.FleetId,
		&gateway.Protocol,
		&gateway.TargetPort,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = errdefs.NewNotFound("gateway not found")
		}
		return
	}
	return gateway, nil
}

func (q Queries) GetGatewayById(ctx context.Context, namespace, id string) (gateway api.Gateway, err error) {
	err = q.db.QueryRow(ctx, baseSelectGateway+" WHERE namespace = $1 AND id = $2", namespace, id).Scan(
		&gateway.Id,
		&gateway.Name,
		&gateway.Namespace,
		&gateway.FleetId,
		&gateway.Protocol,
		&gateway.TargetPort,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = errdefs.NewNotFound("gateway not found")
		}
		return
	}
	return gateway, nil
}

func (q Queries) ListGateways(ctx context.Context, namespace string) ([]api.Gateway, error) {
	rows, err := q.db.Query(ctx, baseSelectGateway+" WHERE namespace = $1", namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gateways := []api.Gateway{}
	for rows.Next() {
		var gateway api.Gateway
		err := rows.Scan(
			&gateway.Id,
			&gateway.Name,
			&gateway.Namespace,
			&gateway.FleetId,
			&gateway.Protocol,
			&gateway.TargetPort,
		)
		if err != nil {
			return nil, err
		}
		gateways = append(gateways, gateway)
	}
	return gateways, nil
}

func (q Queries) CreateGateway(ctx context.Context, gateway api.Gateway) error {
	_, err := q.db.Exec(ctx, `INSERT INTO gateways (id, name, namespace, fleet_id, protocol, target_port) VALUES ($1, $2, $3, $4, $5, $6)`,
		gateway.Id,
		gateway.Name,
		gateway.Namespace,
		gateway.FleetId,
		gateway.Protocol,
		gateway.TargetPort,
	)
	if err != nil {
		return err
	}

	return nil
}

func (q Queries) DeleteGateway(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, `DELETE FROM gateways WHERE  id = $1`, id)
	if err != nil {
		return err
	}
	return nil
}

func (q Queries) UpdateGatewayName(ctx context.Context, id string, name string) error {
	_, err := q.db.Exec(ctx, `UPDATE gateways SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return err
	}
	return nil
}
