package orchestrator

import (
	"context"

	"github.com/valyentdev/ravel/api"
)

func (m *Orchestrator) ListNodes(ctx context.Context) ([]api.Node, error) {
	return m.clusterState.ListNodes(ctx)
}

func (m *Orchestrator) GetNode(ctx context.Context, id string) (api.Node, error) {
	return m.clusterState.GetNode(ctx, id)
}

func (m *Orchestrator) ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error) {
	return m.clusterState.ListNodesInRegion(ctx, region)
}
