package daemon

import (
	"github.com/alexisbouchez/ravel/core/config"
	"github.com/alexisbouchez/ravel/raveld"
	"github.com/spf13/cobra"
)

func NewDaemonCmd() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Start the Ravel daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := config.ReadFile(configFile)
			if err != nil {
				return err
			}

			daemon, err := raveld.NewDaemon(config)
			if err != nil {
				return err
			}

			err = daemon.Start()
			if err != nil {
				return err
			}

			daemon.Run(cmd.Context())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "/etc/ravel/config.toml", "Path to the configuration file")

	return cmd
}
