package orchestrator

import (
	"net/http"

	"github.com/nats-io/nats.go"
	"github.com/valyentdev/ravel/internal/cluster"
	"github.com/valyentdev/ravel/internal/placement"
)

func New(nc *nats.Conn, clusterState *cluster.ClusterState) *Orchestrator {
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
	clusterState *cluster.ClusterState
	nc           *nats.Conn
	broker       *placement.Broker
}
