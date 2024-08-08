package server

import "github.com/spf13/cobra"

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage an API server node",
	Long:  `Commands to manage the ravel api server node.`,
}
