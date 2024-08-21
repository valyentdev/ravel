package orchestrator

import (
	"context"

	"github.com/valyentdev/ravel/pkg/core"
)

func (m *Orchestrator) ListNodes(ctx context.Context) ([]core.Node, error) {
	return m.clusterState.ListNodes(ctx)
}

func (m *Orchestrator) GetNode(ctx context.Context, id string) (core.Node, error) {
	return m.clusterState.GetNode(ctx, id)
}

func (m *Orchestrator) ListNodesInRegion(ctx context.Context, region string) ([]core.Node, error) {
	return m.clusterState.ListNodesInRegion(ctx, region)
}
