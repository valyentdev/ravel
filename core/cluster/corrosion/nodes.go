package corrosion

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (m *Queries) listNodes(ctx context.Context, where string, params ...any) ([]api.Node, error) {
	rows, err := m.dbtx.Query(ctx, "SELECT id, address, agent_port, http_proxy_port, region, heartbeated_at FROM nodes "+where, params...)
	if err != nil {
		return nil, err
	}

	nodes := []api.Node{}
	for rows.Next() {
		var node api.Node
		heartbeatedAt := int64(0)

		err := rows.Scan(&node.Id, &node.Address, &node.AgentPort, &node.HttpProxyPort, &node.Region, &heartbeatedAt)
		if err != nil {
			return nil, err
		}

		node.HeartbeatedAt = time.Unix(heartbeatedAt, 0)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (m *Queries) ListNodes(ctx context.Context) ([]api.Node, error) {
	return m.listNodes(ctx, "")
}

func (m *Queries) ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error) {
	return m.listNodes(ctx, "WHERE region = $1", region)
}

func (m *Queries) GetNode(ctx context.Context, id string) (api.Node, error) {
	node := api.Node{}
	slog.Info("Getting node", "id", id)
	row := m.dbtx.QueryRow(ctx,
		`SELECT id, address, agent_port, http_proxy_port, region, heartbeated_at
				FROM nodes
				WHERE id = $1`,
		id,
	)

	var heartbeatedAt int64

	err := row.Scan(&node.Id, &node.Address, &node.AgentPort, &node.HttpProxyPort, &node.Region, &heartbeatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return node, errdefs.NewNotFound("node not found")
		}
		return node, err
	}

	node.HeartbeatedAt = time.Unix(heartbeatedAt, 0)

	return node, nil
}

func (m *Queries) UpsertNode(ctx context.Context, node api.Node) error {
	_, err := m.dbtx.Exec(ctx,
		`INSERT INTO nodes (id, address, agent_port, http_proxy_port, region, heartbeated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO UPDATE
				SET address = $2, agent_port = $3, http_proxy_port = $4, region = $5, heartbeated_at = $6`,
		node.Id, node.Address, node.AgentPort, node.HttpProxyPort, node.Region, node.HeartbeatedAt.Unix(),
	)
	if err != nil {
		return err
	}

	return nil
}
