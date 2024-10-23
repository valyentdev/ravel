package db

import "github.com/spf13/cobra"

func NewDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
	}

	cmd.AddCommand(NewDBInfosCmd())
	cmd.AddCommand(NewMigrateCmd())

	return cmd
}
