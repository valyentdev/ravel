package agent

import "github.com/spf13/cobra"

func NewAgentCmd() *cobra.Command {
	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage the ravel local agent",
		Long:  `Commands to manage the ravel local agent.`,
	}

	agentCmd.AddCommand(newStartAgentCmd())

	return agentCmd
}
