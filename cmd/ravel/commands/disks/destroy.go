package disks

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newDestroyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "destroy <id>",
		Short: "Destroy a disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDestroyDisk(cmd, args[0])
		},
	}
}

func runDestroyDisk(cmd *cobra.Command, id string) error {
	err := util.GetDaemonClient(cmd).DestroyDisk(cmd.Context(), id)
	if err != nil {
		return err
	}
	cmd.Printf("Disk %s destroyed\n", id)
	return nil
}
