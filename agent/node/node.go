package node

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
)

type Node struct {
	ctx       context.Context
	cancelCtx context.CancelFunc
	cluster   cluster.ClusterState
	localNode api.Node
}

func (n *Node) Id() string {
	return n.localNode.Id
}

func NewNode(c cluster.ClusterState, node api.Node) *Node {
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
	return n.cluster.UpsertNode(ctx, n.localNode)
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

func (n *Node) ListMembers() ([]api.Node, error) {
	return n.cluster.ListNodes(n.ctx)
}

func (n *Node) GetMember(id string) (api.Node, error) {
	return n.cluster.GetNode(n.ctx, id)
}
