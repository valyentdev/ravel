package agentclient

import (
	"net/http"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/internal/httpclient"
)

type AgentClient struct {
	baseUrl    string
	httpClient *http.Client
	client     *httpclient.Client
}

var _ structs.Agent = (*AgentClient)(nil)

func NewAgentClient(c *http.Client, baseUrl string) *AgentClient {
	return &AgentClient{baseUrl: baseUrl, client: httpclient.NewClient(baseUrl, c), httpClient: c}
}
