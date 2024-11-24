package instance

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
)

func newListInstancesCmd() *cobra.Command {

	var lsCmd = &cobra.Command{
		Use:   "ls",
		Short: "List all  instances",
		Long:  `List all instances`,
		RunE:  runListInstances,
	}

	return lsCmd

}

func runListInstances(cmd *cobra.Command, args []string) error {
	instances, err := util.GetDaemonClient(cmd).ListInstances(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list instances: %w", err)
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "ID\tSTATUS\tIMAGE\tLOCAL IP")
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", instance.Id, instance.State.Status, instance.ImageRef, instance.Network.Local.InstanceIP)
	}
	w.Flush()

	return nil
}
