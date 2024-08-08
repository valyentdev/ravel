package instance

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	raveldclient "github.com/valyentdev/ravel/cmd/ravel/client"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/proto"
	"sigs.k8s.io/yaml"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new instance",
	Long: `Create & start a new instance from a given image
The instance spec is defined in a json or yaml file.`,
	Run: func(cmd *cobra.Command, args []string) {
		var instanceId string

		if len(args) > 1 {
			instanceId = args[0]
		} else {
			instanceId = id.Generate()
		}

		configFile := cmd.Flag("config").Value.String()
		file, err := os.ReadFile(configFile)
		if err != nil {
			cmd.Println("Unable to read config file ", configFile, "\nerror: ", err)
			return
		}

		var config core.MachineConfig
		if err = yaml.Unmarshal(file, &config); err != nil {
			cmd.Println("Unable to unmarshal config file ", configFile, "\nerror: ", err)
			return
		}

		res, err := raveldclient.DaemonClient.CreateInstance(context.Background(), &proto.CreateInstanceRequest{
			MachineId: instanceId,
			Namespace: "local",
			Config:    core.MachineConfigToProto(config),
		})

		if err != nil {
			cmd.Println("Unable to create instance: ", err)
			return
		}

		cmd.Println(res.InstanceId)

	},
}

func init() {
	createCmd.Flags().StringP("config", "c", "", "Config file which contains instance spec")
	createCmd.MarkFlagRequired("config")

	// If config file is not specified, we need to specify all the flags

	InstanceCmd.AddCommand(createCmd)
}
