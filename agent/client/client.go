package agentclient

import (
	"net/http"

	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/internal/httpclient"
)

type AgentClient struct {
	baseUrl string
	client  *httpclient.Client
}

var _ cluster.Agent = (*AgentClient)(nil)

func NewAgentClient(c *http.Client, baseUrl string) *AgentClient {
	return &AgentClient{baseUrl: baseUrl, client: httpclient.NewClient(baseUrl, c)}
}
