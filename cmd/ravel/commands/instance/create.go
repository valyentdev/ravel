package instance

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/cmd/ravel/util"
	"github.com/valyentdev/ravel/core/instance"
	"sigs.k8s.io/yaml"
)

type createOptions struct {
	config string
}

func newCreateInstanceCmd() *cobra.Command {
	var opts createOptions

	createCmd := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a new instance",
		Long: `Create & start a new instance from a given image
The instance spec is defined in a json or yaml file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstance(cmd, args[0], opts)
		},
		Args: cobra.ExactArgs(1),
	}

	createCmd.Flags().StringVarP(&opts.config, "config", "c", "", "Config file which contains instance spec")
	createCmd.MarkFlagRequired("config")

	return createCmd
}

func createInstance(cmd *cobra.Command, id string, opt createOptions) error {
	file, err := os.ReadFile(opt.config)
	if err != nil {
		return fmt.Errorf("unable to read config file %s: %w", opt.config, err)
	}

	var config instance.InstanceConfig
	if err = yaml.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("unable to unmarshal config file %s: %w", opt.config, err)
	}

	res, err := util.GetAgentClient(cmd).CreateInstance(cmd.Context(), structs.InstanceOptions{
		Id:     id,
		Config: config,
	})

	if err != nil {
		return fmt.Errorf("unable to create instance: %w", err)
	}

	fmt.Println(res.Id)
	return nil
}
