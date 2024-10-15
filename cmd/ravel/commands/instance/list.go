package instance

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
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
	instances, err := GetClient(cmd).ListInstances(context.Background())
	if err != nil {
		return fmt.Errorf("unable to list instances: %w", err)
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)

	fmt.Fprintln(w, "ID\tSTATUS\tIMAGE\tLOCAL IP")
	for _, instance := range instances {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", instance.Id, instance.State.Status, instance.Config.Workload.Image, instance.LocalIPV4)
	}
	w.Flush()

	return nil
}
