package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/cmd/agent"
	"github.com/valyentdev/ravel/cmd/ravel/cmd/instance"
	"github.com/valyentdev/ravel/cmd/ravel/cmd/server"
)

var rootCmd = &cobra.Command{
	Use:   "ravel",
	Short: "A cli tool for managing raveld.",
	Long:  ``,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(instance.InstanceCmd)
	rootCmd.AddCommand(agent.AgentCmd)
	rootCmd.AddCommand(server.ServerCmd)
}
