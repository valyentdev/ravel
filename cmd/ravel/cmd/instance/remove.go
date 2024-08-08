package instance

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/pkg/proto"
)

var removeCmd = &cobra.Command{

	Use:     "remove",
	Aliases: []string{"rm"},
	Short:   "Remove a instance",
	Long:    `Remove a instance. The instance must be stopped.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("Please specify a instance ID")
			return
		}

		instanceId := args[0]

		_, err := raveldclient.DaemonClient.DestroyInstance(context.Background(), &proto.DestroyInstanceRequest{Id: instanceId})

		if err != nil {
			cmd.Println("Unable to remove instance: ", err)
			os.Exit(1)
		}

		cmd.Println("Instance removed")

	},
}

func init() {
	InstanceCmd.AddCommand(removeCmd)
}
