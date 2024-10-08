package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/valyentdev/ravel/cmd/ravel/commands"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	rootCmd := commands.NewRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

}
