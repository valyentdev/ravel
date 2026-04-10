package commands

import (
	"log/slog"

	"github.com/alexisbouchez/ravel/cmd/ravel/commands/ctl"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/daemon"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/db"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/disks"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/images"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/instance"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/server"
	"github.com/alexisbouchez/ravel/cmd/ravel/commands/tls"
	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(
		instance.NewInstanceCmd(),
		daemon.NewDaemonCmd(),
		server.NewServerCmd(),
		db.NewDBCmd(),
		images.NewImagesCmd(),
		tls.NewTLSCommand(),
		disks.NewDisksCmd(),
		ctl.NewCtlCmd(),
	)

	return rootCmd
}
