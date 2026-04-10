package disks

import "github.com/spf13/cobra"

func NewDisksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disks",
		Short: "Manage disks",
	}

	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newDestroyCmd())

	return cmd
}
