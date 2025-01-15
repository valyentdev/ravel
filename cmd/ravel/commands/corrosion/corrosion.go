package corrosion

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/core/cluster/corrosion"
	"github.com/valyentdev/ravel/core/config"
)

func NewCorroCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "corrosion",
		Aliases: []string{"corro"},
		Short:   "Commands to interact with the Corrosion instance",
	}

	cmd.AddCommand(newMigrateCmd())
	return cmd
}

func newMigrateCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the corrosion database to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := config.ReadFile(configPath)
			if err != nil {
				return err
			}

			err = corrosion.RunCorrosionMigrations(cmd.Context(), config.Corrosion.Config())
			if err != nil {
				return err
			}

			cmd.Println("Corrosion database migrated successfully")

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "/etc/ravel/config.toml", "Path to the configuration file")

	return cmd
}
