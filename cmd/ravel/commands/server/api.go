package server

import "github.com/spf13/cobra"

func NewServerCmd() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Manage an API server node",
		Long:  `Commands to manage the ravel api server node.`,
	}

	serverCmd.AddCommand(newStartApiServerCmd())

	return serverCmd
}
