package agent

import "github.com/spf13/cobra"

var AgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage the ravel local agent",
	Long:  `Commands to manage the ravel local agent.`,
}
