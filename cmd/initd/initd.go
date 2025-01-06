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

	go env.Run()

	if err := api.ServeInitdAPI(env); err != nil {
		slog.Error("[ravel-initd] Failed to start API server: %v", "err", err)
	}
}
