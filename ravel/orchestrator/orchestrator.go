package orchestrator

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	agentclient "github.com/alexisbouchez/ravel/agent/client"
	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/core/cluster/placement"
	"github.com/nats-io/nats.go"
)

func New(nc *nats.Conn, clusterState cluster.ClusterState, tlsConfig *tls.Config) *Orchestrator {
	broker := placement.NewBroker(nc)
	httpClient := http.Client{ // to be tuned in the future
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return &Orchestrator{
		tlsEnabled:   tlsConfig != nil,
		httpClient:   &httpClient,
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
	tlsEnabled   bool
}

func (o *Orchestrator) getAgentClient(node string) (*agentclient.AgentClient, error) {
	member, err := o.clusterState.GetNode(context.Background(), node)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	var scheme string
	if o.tlsEnabled {
		scheme = "https"
	} else {
		scheme = "http"
	}

	client := agentclient.NewAgentClient(o.httpClient, fmt.Sprintf("%s://%s", scheme, member.AgentAddress()))

	return client, nil
}

// GetAgentClient returns an agent client for the specified node
func (o *Orchestrator) GetAgentClient(nodeId string) (*agentclient.AgentClient, error) {
	return o.getAgentClient(nodeId)
}
