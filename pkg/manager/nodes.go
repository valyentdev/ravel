package manager

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

func (m *Manager) ListNodes(ctx context.Context) ([]core.Node, error) {
	return m.clusterState.ListNodes(ctx)
}

func (m *Manager) GetNode(ctx context.Context, id string) (core.Node, error) {
	return m.clusterState.GetNode(ctx, id)
}

func (m *Manager) ListNodesInRegion(ctx context.Context, region string) ([]core.Node, error) {
	return m.clusterState.ListNodesInRegion(ctx, region)
}
