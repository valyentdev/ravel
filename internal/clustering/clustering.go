package clustering

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/pkg/core"
)

type Node struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	cluster   *cluster.ClusterState
	localNode core.Node
}

func NewNode(c *cluster.ClusterState, node core.Node) *Node {
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		cluster:   c,
		localNode: node,
		ctx:       ctx,
		cancelCtx: cancel,
	}
}

func (n *Node) Start() error {
	go n.startHeartbeating(n.ctx)

	return nil
}

func (n *Node) Stop() {
	n.cancelCtx()
}

func (n *Node) heartbeat(ctx context.Context) error {
	n.localNode.HeartbeatedAt = time.Now()
	n.cluster.UpsertNode(ctx, n.localNode)
	return nil
}

func (n *Node) startHeartbeating(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := n.heartbeat(context.Background()); err != nil {
				slog.Error("failed to heartbeat", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) ListMembers() ([]core.Node, error) {
	return n.cluster.ListNodes(n.ctx)
}

func (n *Node) GetMember(id string) (core.Node, error) {
	return n.cluster.GetNode(n.ctx, id)
}
