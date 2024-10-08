package main

import (
	"log/slog"
	"net/http"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/internal/proxy"
)

func main() {
	corro := corroclient.NewCorroClient(corroclient.Config{
		URL: "http://localhost:8081",
	})

	slog.Info("Starting ravel proxy")
	ravelProxy := proxy.NewRavelProxy(corro)

	ravelProxy.Start()

	http.ListenAndServe(":3001", http.HandlerFunc(ravelProxy.Handle))
}
