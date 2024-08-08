package instance

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/pkg/proto"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running instance",
	Long:  `Stop a running instance from a given config file`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("Please specify a instance ID")
			return
		}

		instanceId := args[0]

		_, err := raveldclient.DaemonClient.StopInstance(context.Background(), &proto.StopInstanceRequest{Id: instanceId})

		if err != nil {
			cmd.Println("Unable to stop instance: ", err)
			os.Exit(1)
		}

		cmd.Println("Instance stopped")

	},
}

func init() {
	InstanceCmd.AddCommand(stopCmd)
}
