package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/internal/server"
	"github.com/valyentdev/ravel/pkg/core/config"
)

type startApiServerOptions struct {
	config string
}

func newStartApiServerCmd() *cobra.Command {
	var opts startApiServerOptions

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the API server",
		Long:  `Start the API server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startApiServer(cmd.Context(), opts)
		},
	}

	startCmd.Flags().StringVarP(&opts.config, "config", "c", "ravel.json", "Path to config file")
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
	waitShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		slog.Info("Shutting down server")
		err := s.Shutdown(context.Background())
		if err != nil {
			slog.Error("failed to shutdown server", "error", err)
		}
		close(waitShutdown)
	}()

	err = s.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	<-waitShutdown

	return nil
}
