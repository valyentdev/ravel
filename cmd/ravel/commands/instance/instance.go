package instance

import (
	"github.com/spf13/cobra"
)

func NewInstanceCmd() *cobra.Command {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "Manage ravel instances",
		Long:  `Commands to manage ravel instances.`,
	}

	instanceCmd.AddCommand(newCreateInstanceCmd())
	instanceCmd.AddCommand(newInstanceExec())
	instanceCmd.AddCommand(newListInstancesCmd())
	instanceCmd.AddCommand(newDestroyInstanceCmd())
	instanceCmd.AddCommand(newStartInstanceCmd())
	instanceCmd.AddCommand(newStopCmd())

	return instanceCmd
}
