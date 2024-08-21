package commands

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/commands/instance"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ravel",
		Short: "A cli tool for managing raveld.",
		Long:  ``,
	}

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.AddCommand(instance.NewInstanceCmd())

	return rootCmd
}
