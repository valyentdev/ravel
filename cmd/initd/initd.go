package main

import (
	"log/slog"

	"github.com/valyentdev/ravel/initd/api"
	"github.com/valyentdev/ravel/initd/environment"
)

func main() {
	env := &environment.Env{}

	if err := env.Init(); err != nil {
		panic(err)
	}

	err := env.Start()
	if err != nil {
		slog.Error("[ravel-initd] Failed to start initd: %v", "err", err)
	}

	// even if we fail to start initd, we should still start the API server so ravel can notice quickly
	if err := api.ServeInitdAPI(env); err != nil {
		slog.Error("[ravel-initd] Failed to start API server: %v", "err", err)
	}
}
