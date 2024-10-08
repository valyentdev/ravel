package server

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/internal/server"
	"github.com/valyentdev/ravel/pkg/core/config"
)

type startApiServerOptions struct {
	env    string
	config string
}

func newStartApiServerCmd() *cobra.Command {
	var opts startApiServerOptions

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the API server",
		Long:  `Start the API server.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startApiServer(opts)
		},
	}

	startCmd.Flags().StringVar(&opts.env, "env", ".env", "Path to .env file")
	startCmd.Flags().StringVarP(&opts.config, "config", "c", "ravel.json", "Path to config file")
	return startCmd
}

func startApiServer(opt startApiServerOptions) error {
	slog.Info("Starting API server")
	err := godotenv.Load(opt.env)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ravelConfig, err := config.ReadFile(opt.config)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	s, err := server.NewServer(ravelConfig)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	err = s.Serve()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
