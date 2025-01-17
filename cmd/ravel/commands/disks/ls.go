package disks

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List disks",
		RunE: func(cmd *cobra.Command, args []string) error {
			disks, err := util.GetDaemonClient(cmd).ListDisks(cmd.Context())
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)

			fmt.Fprintln(w, "ID\tINSTANCE\tSIZE MB\tCREATED AT")
			for _, disk := range disks {
				var instance string
				if disk.AttachedInstance != "" {
					instance = disk.AttachedInstance
				} else {
					instance = "-"
				}

				fmt.Fprintf(w, "%s\t%s\t%dMB\t%s\n", disk.Id, instance, disk.SizeMB, disk.CreatedAt)
			}
			w.Flush()

			return nil
		},
	}
}
