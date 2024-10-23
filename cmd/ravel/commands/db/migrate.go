package db

import (
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/pkg/core/config"
	"github.com/valyentdev/ravel/pkg/ravel/db/schema"
)

type migrateOpts struct {
	config string
	to     string
}

func parseVersion(version string) (int, error) {
	v, err := strconv.Atoi(version)
	if err != nil {
		return 0, fmt.Errorf("invalid version: %w", err)
	}

	if v < 0 {
		return 0, fmt.Errorf("invalid version: %d", v)
	}

	return v, nil
}

func NewMigrateCmd() *cobra.Command {
	var migrateOptions migrateOpts

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate the database",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrateCmd(cmd, migrateOptions)
		},
	}

	cmd.Flags().StringVarP(&migrateOptions.config, "config", "c", "/etc/ravel/config.json", "Path to the configuration file")
	cmd.Flags().StringVar(&migrateOptions.to, "to", "", "Migrate to a specific version")
	cmd.MarkFlagRequired("to")
	return cmd
}

func runMigrateCmd(cmd *cobra.Command, opt migrateOpts) error {
	config, err := config.ReadFile(opt.config)
	if err != nil {
		return err
	}

	conn, err := pgx.Connect(cmd.Context(), config.PostgresURL)
	if err != nil {
		return err
	}

	migrator, err := schema.NewMigrator(conn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if opt.to == "latest" {
		err = migrator.Migrate(cmd.Context())
		if err != nil {
			return fmt.Errorf("failed to migrate: %w", err)
		}
		fmt.Println("Database migrated to the latest version")
		return err
	} else {
		version, err := parseVersion(opt.to)
		if err != nil {
			return err
		}

		err = migrator.MigrateTo(cmd.Context(), int32(version))
		if err != nil {
			return fmt.Errorf("failed to migrate to version %d: %w", version, err)
		}

		fmt.Printf("Database migrated to version %d\n", version)
	}

	return nil
}
