package main

import (
	"log/slog"
	"os"

	"github.com/valyentdev/ravel/core/jailer"
)

func main() {
	if err := jailer.Run(); err != nil {
		slog.Error("jailer run failed", "error", err)
		os.Exit(1)
	}
}
