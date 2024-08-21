package instance

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/pkg/core"
	"sigs.k8s.io/yaml"
)

type createOptions struct {
	config string
}

func newCreateInstanceCmd() *cobra.Command {
	var opts createOptions

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new instance",
		Long: `Create & start a new instance from a given image
The instance spec is defined in a json or yaml file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstance(cmd, opts)
		},
	}

	createCmd.Flags().StringVarP(&opts.config, "config", "c", "", "Config file which contains instance spec")
	createCmd.MarkFlagRequired("config")

	return createCmd
}

func createInstance(cmd *cobra.Command, opt createOptions) error {
	var instanceId string

	file, err := os.ReadFile(opt.config)
	if err != nil {
		return fmt.Errorf("unable to read config file %s: %w", opt.config, err)
	}

	var config core.InstanceConfig
	if err = yaml.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("unable to unmarshal config file %s: %w", opt.config, err)
	}

	res, err := GetClient(cmd).CreateInstance(cmd.Context(), core.CreateInstancePayload{
		MachineId: instanceId,
		Namespace: "local",
		Config:    config,
	})

	if err != nil {
		return fmt.Errorf("unable to create instance: %w", err)
	}

	fmt.Println(res.Id)
	return nil
}
