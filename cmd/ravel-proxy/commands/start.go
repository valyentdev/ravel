package commands

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/edge"
	"github.com/valyentdev/ravel/proxy/local"
	"github.com/valyentdev/ravel/proxy/server"
)

func newStartCmd() *cobra.Command {
	var configPath string
	var mode string

	edge := cobra.Command{
		Use:   "start",
		Short: "Start the proxy",

		RunE: func(cmd *cobra.Command, args []string) error {
			if mode != string(proxy.Edge) && mode != string(proxy.Local) {
				return fmt.Errorf("invalid mode: %s", mode)
			}

			return runStart(cmd.Context(), proxy.Mode(mode), configPath)
		},
	}

	edge.Flags().StringVarP(&mode, "mode", "m", "", "Mode of the proxy (edge, local)")
	edge.MarkFlagRequired("mode")
	edge.Flags().StringVarP(&configPath, "config", "c", "/etc/ravel/proxy.toml", "Path to the config file")

	return &edge
}

func runStart(ctx context.Context, mode proxy.Mode, configPath string) error {
	config, err := proxy.ReadConfigFile(configPath)
	if err != nil {
		return err
	}

	slog.Info("Starting proxy in", "mode", mode)
	switch mode {
	case proxy.Edge:
		return runEdge(ctx, config)
	case proxy.Local:
		return runLocal(ctx, config)
	}

	return nil
}

func runEdge(ctx context.Context, config *proxy.Config) error {
	edgeProxy, err := edge.NewEdgeProxyServer(config)
	if err != nil {
		return err
	}
	return runServer(ctx, edgeProxy, 60*time.Second)
}

func runLocal(ctx context.Context, config *proxy.Config) error {
	localProxy := local.NewLocalProxyServer(config)
	return runServer(ctx, localProxy, 60*time.Second)
}

func runServer(ctx context.Context, server *server.Server, shutdownTimeout time.Duration) error {
	if err := server.Start(); err != nil {
		return err
	}

	<-ctx.Done()
	slog.Info("Shutting down server")

	shutdownCTX, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	server.Shutdown(shutdownCTX)

	return nil
}
