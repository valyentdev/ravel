package commands

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/raveld/commands/agent"
	"github.com/valyentdev/ravel/cmd/raveld/commands/server"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raveld",
		Short: "Ravel daemon",
	}

	cmd.AddCommand(agent.NewAgentCmd())
	cmd.AddCommand(server.NewServerCmd())

	return cmd
}
