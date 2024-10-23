package commands

import (
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/commands/agent"
	"github.com/valyentdev/ravel/cmd/ravel/commands/db"
	"github.com/valyentdev/ravel/cmd/ravel/commands/instance"
	"github.com/valyentdev/ravel/cmd/ravel/commands/server"
)

func NewRootCmd() *cobra.Command {
	var opt struct {
		debug bool
	}

	rootCmd := &cobra.Command{
		Use:   "ravel",
		Short: "A cli tool for managing raveld.",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if opt.debug {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				slog.Debug("Debug mode enabled")
			}
		},
	}

	rootCmd.PersistentFlags().BoolVar(&opt.debug, "debug", false, "Enable debug logging")
	rootCmd.AddCommand(instance.NewInstanceCmd())
	rootCmd.AddCommand(agent.NewAgentCmd())
	rootCmd.AddCommand(server.NewServerCmd())
	rootCmd.AddCommand(db.NewDBCmd())

	return rootCmd
}
