package corrosion

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/errdefs"
)

func (m *CorrosionClusterState) listNodes(ctx context.Context, where string, params ...any) ([]api.Node, error) {
	rows, err := m.corroclient.Query(ctx, corroclient.Statement{Query: "SELECT id, address, agent_port, http_proxy_port, region, heartbeated_at FROM nodes " + where, Params: params})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []api.Node{}, nil
		}
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

func (m *CorrosionClusterState) ListNodes(ctx context.Context) ([]api.Node, error) {
	return m.listNodes(ctx, "")
}

func (m *CorrosionClusterState) ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error) {
	return m.listNodes(ctx, "WHERE region = $1", region)
}

func (m *CorrosionClusterState) GetNode(ctx context.Context, id string) (api.Node, error) {
	node := api.Node{}
	slog.Info("Getting node", "id", id)
	row, err := m.corroclient.QueryRow(ctx, corroclient.Statement{
		Query: `SELECT id, address, agent_port, http_proxy_port, region, heartbeated_at
				FROM nodes
				WHERE id = $1`,
		Params: []interface{}{id},
	})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return node, errdefs.NewNotFound("node not found")
		}
		return node, err
	}

	var heartbeatedAt int64

	err = row.Scan(&node.Id, &node.Address, &node.AgentPort, &node.HttpProxyPort, &node.Region, &heartbeatedAt)
	if err != nil {
		return node, err
	}

	node.HeartbeatedAt = time.Unix(heartbeatedAt, 0)

	return node, nil
}

func (m *CorrosionClusterState) UpsertNode(ctx context.Context, node api.Node) error {
	results, err := m.corroclient.Exec(ctx, []corroclient.Statement{{
		Query: `INSERT INTO nodes (id, address, agent_port, http_proxy_port, region, heartbeated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO UPDATE
				SET address = $2, agent_port = $3, http_proxy_port = $4, region = $5, heartbeated_at = $6`,
		Params: []interface{}{node.Id, node.Address, node.AgentPort, node.HttpProxyPort, node.Region, node.HeartbeatedAt.Unix()},
	}})
	if err != nil {
		return err
	}

	if results.Results[0].Err() != nil {
		return results.Results[0].Err()
	}

	return err
}