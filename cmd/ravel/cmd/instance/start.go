package instance

import (
	"context"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/pkg/proto"
)

var startCmd = &cobra.Command{
	Use:                   "start",
	Short:                 "Start a previously stopped instance",
	Long:                  `Start a previously stopped instance. This is a no-op if the instance is already running.`,
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Println("Please specify a instanceId")
			cmd.Help()
			return
		}

		instanceId := args[0]

		_, err := raveldclient.DaemonClient.StartInstance(context.Background(), &proto.StartInstanceRequest{Id: instanceId})

		if err != nil {
			cmd.Println("Error while starting instance: ", err)
			return
		}

		cmd.Println(instanceId)

	},
}

func init() {
	InstanceCmd.AddCommand(startCmd)

}
