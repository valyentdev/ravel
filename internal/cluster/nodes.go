package cluster

import (
	"context"
	"time"

	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/helper/corroclient"
	"github.com/valyentdev/ravel/pkg/ravelerrors"
)

func (m *ClusterState) listNodes(ctx context.Context, where string, params ...any) ([]core.Node, error) {
	rows, err := m.corroclient.QueryContext(ctx, corroclient.Statement{Query: "SELECT id, address, region, heartbeated_at FROM nodes " + where, Params: params})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return []core.Node{}, nil
		}
		return nil, err
	}

	nodes := []core.Node{}

	for rows.Next() {
		var node core.Node
		heartbeatedAt := int64(0)

		err := rows.Scan(&node.Id, &node.Address, &node.Region, &heartbeatedAt)
		if err != nil {
			return nil, err
		}

		node.HeartbeatedAt = time.Unix(heartbeatedAt, 0)
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (m *ClusterState) ListNodes(ctx context.Context) ([]core.Node, error) {
	return m.listNodes(ctx, "")
}

func (m *ClusterState) ListNodesInRegion(ctx context.Context, region string) ([]core.Node, error) {
	return m.listNodes(ctx, "WHERE region = $1", region)
}

func (m *ClusterState) GetNode(ctx context.Context, id string) (core.Node, error) {
	node := core.Node{}
	row, err := m.corroclient.QueryRowContext(ctx, corroclient.Statement{
		Query: `SELECT id, address, region, heartbeated_at
				FROM nodes
				WHERE id = $1`,
		Params: []interface{}{id},
	})
	if err != nil {
		if err == corroclient.ErrNoRows {
			return node, ravelerrors.NewNotFound("node not found")
		}
		return node, err
	}

	var heartbeatedAt int64

	err = row.Scan(&node.Id, &node.Address, &node.Region, &heartbeatedAt)
	if err != nil {
		return node, err
	}

	node.HeartbeatedAt = time.Unix(heartbeatedAt, 0)

	return node, nil
}

func (m *ClusterState) UpsertNode(ctx context.Context, node core.Node) error {
	results, err := m.corroclient.ExecContext(ctx, []corroclient.Statement{{
		Query: `INSERT INTO nodes (id, address, region, heartbeated_at)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (id) DO UPDATE
				SET address = $2, region = $3, heartbeated_at = $4`,
		Params: []interface{}{node.Id, node.Address, node.Region, node.HeartbeatedAt.Unix()},
	}})
	if err != nil {
		return err
	}

	if results.Results[0].Err() != nil {
		return results.Results[0].Err()
	}

	return err
}
