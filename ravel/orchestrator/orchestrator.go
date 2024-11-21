package orchestrator

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nats-io/nats.go"
	agentclient "github.com/valyentdev/ravel/agent/client"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/cluster/placement"
)

func New(nc *nats.Conn, clusterState cluster.ClusterState) *Orchestrator {
	broker := placement.NewBroker(nc)

	return &Orchestrator{
		httpClient:   http.DefaultClient, // to be tuned in the future
		nc:           nc,
		clusterState: clusterState,
		broker:       broker,
	}
}

type Orchestrator struct {
	httpClient   *http.Client // used to communicate with agents
	clusterState cluster.ClusterState
	nc           *nats.Conn
	broker       *placement.Broker
}

func (m *Orchestrator) getAgentClient(node string) (*agentclient.AgentClient, error) {
	member, err := m.clusterState.GetNode(context.Background(), node)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	client := agentclient.NewAgentClient(m.httpClient, fmt.Sprintf("http://%s", member.AgentAddress()))

	return client, nil
}
