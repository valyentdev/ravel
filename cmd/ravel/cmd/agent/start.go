package agent

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/internal/agent"
	"github.com/valyentdev/ravel/pkg/config"
	"github.com/valyentdev/ravel/pkg/proto"
)

func init() {
	startCmd.Flags().StringP("config", "c", "ravel.json", "Path to the configuration file")
	startCmd.Flags().Bool("debug", false, "Enable debug logging")
	AgentCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		debugEnabled, _ := cmd.Flags().GetBool("debug")
		if debugEnabled {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		configFile := cmd.Flag("config").Value.String()

		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			slog.Error("config file does not exist")
			os.Exit(1)
		}

		ravelConfig, err := config.ReadFile(configFile)
		if err != nil {
			slog.Error("failed to read config file", "error", err)
		}

		a, err := agent.New(ravelConfig)
		if err != nil {
			slog.Error("failed to create agent", "error", err)
			os.Exit(1)
		}

		err = a.Start(ctx)
		if err != nil {
			slog.Error("agent exited with error", "error", err)
			os.Exit(1)
		}

		listener, err := net.Listen("tcp", ravelConfig.Agent.Address)
		if err != nil {
			panic(err)
		}

		grpcserver := grpc.NewServer()

		proto.RegisterAgentServiceServer(grpcserver, a)

		go func() {
			<-ctx.Done()

			grpcserver.GracefulStop()

			err = a.Stop()
			if err != nil {
				slog.Error("agent stopped with error", "error", err)
			}
		}()

		grpcserver.Serve(listener)
	},
}
