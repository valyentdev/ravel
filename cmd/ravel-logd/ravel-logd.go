package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/valyentdev/ravel/internal/logd"
)

var (
	configPath = flag.String("config", "/etc/ravel/logd.yaml", "Path to the config file")
)

func init() {
	flag.Parse()

}

func main() {

	// Load the config
	config, err := logd.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get logs file from config
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	_, ctx, err = logd.NewLogger(ctx)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	// Create a new LogManager
	logWatcherSvc, err := logd.NewLogWatcherService(config)
	if err != nil {
		log.Fatalf("Failed to create LogManager: %v", err)
	}

	// Start the LogManager
	go logWatcherSvc.Start(ctx)

	// Wait for a signal to stop
	<-ctx.Done()

	// TODO:jnfrati - Add a graceful shutdown
	// Do we want to wait for all loggers to reach EOF?
	// Or do we want to stop all loggers immediately?
}
