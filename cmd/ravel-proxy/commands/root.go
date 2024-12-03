package commands

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var debug bool
	root := cobra.Command{
		Use:   "ravel-proxy",
		Short: "Ravel Proxy",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if debug {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				slog.Debug("Debug mode enabled")
			}
		},
	}

	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	root.AddCommand(newStartCmd())

	return &root
}
