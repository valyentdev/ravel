package db

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/ravel/db/schema"
)

func NewDBInfosCmd() *cobra.Command {

	var infosOptions struct {
		config string
	}

	cmd := &cobra.Command{
		Use:   "infos",
		Short: "Show database informations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBInfosCmd(cmd, infosOptions.config)
		},
	}

	cmd.Flags().StringVarP(&infosOptions.config, "config", "c", "/etc/ravel/config.toml", "Path to the configuration file")

	return cmd
}

func runDBInfosCmd(cmd *cobra.Command, configPath string) error {
	config, err := config.ReadFile(configPath)
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(cmd.Context(), config.Server.PostgresURL)
	if err != nil {
		return err
	}
	defer conn.Close(cmd.Context())

	migrator, err := schema.NewMigrator(conn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	current, total, migrations, err := migrator.Info(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get migration info: %w", err)
	}

	cmd.Printf("Current version: %d\n", current)
	cmd.Printf("Total migrations: %d\n", total)
	cmd.Println("\nMigrations:")
	for i, m := range migrations {
		if i+1 <= int(current) {
			cmd.Print("[✓] ")
		} else {
			cmd.Print("[ ] ")
		}
		cmd.Printf("%d  %s", i+1, m.Name)
		cmd.Println()
	}

	return nil
}
