package agent

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/valyentdev/ravel/pkg/agent"
	"github.com/valyentdev/ravel/pkg/core/config"

	"github.com/spf13/cobra"
)

type agentStartOpt struct {
	configFile string
}

func newStartAgentCmd() *cobra.Command {
	var opt agentStartOpt

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the ravel agent",
		Long:  `Start the ravel agent`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentStart(opt)
		},
	}

	startCmd.Flags().StringVarP(&opt.configFile, "config", "c", "/etc/ravel/ravel.toml", "config file")
	return startCmd
}

func runAgentStart(opt agentStartOpt) error {
	configFile := opt.configFile

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if _, err := os.Stat(configFile); err != nil {
		return fmt.Errorf("cannot find config file: %w", err)
	}

	ravelConfig, err := config.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	a, err := agent.New(ravelConfig)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	err = a.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	server := agent.NewAgentServer(a, fmt.Sprintf("%s:%d", ravelConfig.Agent.Address, ravelConfig.Agent.AgentPort))

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				slog.Info("server shutdown")
				return
			}

			slog.Error("server exited with error", "error", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("failed to gracefully shutdown server before timeout", "error", err)
	}

	return nil
}
