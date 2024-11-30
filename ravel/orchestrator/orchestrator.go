package orchestrator

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/nats-io/nats.go"
	agentclient "github.com/valyentdev/ravel/agent/client"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/cluster/placement"
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
