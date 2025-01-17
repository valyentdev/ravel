package disks

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
	"github.com/valyentdev/ravel/core/daemon"
)

func newCreateCmd() *cobra.Command {
	var size uint64

	cmd := &cobra.Command{
		Use:   "create <id> [--size|-s <size>]",
		Short: "Create a disk",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				cmd.Println("id is required")
				return cmd.Usage()
			}
			return runCreateDisk(cmd, args[0], size)
		},
	}

	cmd.Flags().Uint64VarP(&size, "size", "s", 512, "Size of the disk in MB")

	return cmd
}

func runCreateDisk(cmd *cobra.Command, id string, size uint64) error {
	disk, err := util.GetDaemonClient(cmd).CreateDisk(cmd.Context(), daemon.DiskOptions{
		SizeMB: size,
		Id:     id,
	})

	if err != nil {
		return err
	}

	cmd.Printf("Disk %s created\n", disk.Id)
	return nil
}
