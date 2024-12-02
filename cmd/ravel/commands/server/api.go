package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/ravel/server"
)

type startApiServerOptions struct {
	config string
}

func NewServerCmd() *cobra.Command {
	var opts startApiServerOptions

	var startCmd = &cobra.Command{
		Use:   "server",
		Short: "Start the API server",
		Long:  `Start the API server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startApiServer(cmd.Context(), opts)
		},
	}

	startCmd.Flags().StringVarP(&opts.config, "config", "c", "/etc/ravel/ravel.toml", "Path to config file")
	return startCmd
}

func startApiServer(ctx context.Context, opt startApiServerOptions) error {
	slog.Info("Starting API server")
	ravelConfig, err := config.ReadFile(opt.config)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	s, err := server.NewServer(ravelConfig)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	s.Run(ctx)

	return nil
}
