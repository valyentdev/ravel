package util

import (
	"net/http"

	"github.com/spf13/cobra"
	agentclient "github.com/valyentdev/ravel/agent/client"
)

func GetAgentClient(cmd *cobra.Command) *agentclient.AgentClient {
	return agentclient.NewAgentClient(http.DefaultClient, "http://127.0.0.1:8080")
}
