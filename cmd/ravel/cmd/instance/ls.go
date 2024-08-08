package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/pkg/proto"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all  instances",
	Long:  `List all instances`,
	Run: func(cmd *cobra.Command, args []string) {
		res, err := raveldclient.DaemonClient.ListInstances(context.Background(), &proto.ListInstancesRequest{})
		if err != nil {
			cmd.PrintErrln("Unable to list instances: ", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 1, 1, 1, ' ', 0)

		fmt.Fprintln(w, "ID\tSTATUS\tIMAGE\tLOCAL IP")
		for _, instance := range res.Instances {
			json, err := json.Marshal(instance)
			if err != nil {
				cmd.PrintErrln("Unable to marshal instance: ", err)
				os.Exit(1)
			}

			fmt.Println(string(json))
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", instance.Id, instance.State, instance.Config.Workload.Image, instance.LocalIpv4)
		}
		w.Flush()
	},
}

func init() {
	InstanceCmd.AddCommand(lsCmd)
}
