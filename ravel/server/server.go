package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/ravel"
	"github.com/valyentdev/ravel/ravel/server/endpoints"
)

type Server struct {
	ravel  *ravel.Ravel
	server *http.Server
}

func NewServer(c config.RavelConfig) (*Server, error) {
	r, err := ravel.New(c)
	if err != nil {
		return nil, err
	}

	address := c.Server.Address
	if address == "" {
		address = ":3000"
	}

	errdefs.OverrideHumaErrorBuilder()

	mux := http.NewServeMux()

	humaConfig := getHumaConfig()
	api := humago.New(mux, humaConfig)
	e := endpoints.New(r)
	e.Register(api)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return &Server{
		ravel:  r,
		server: server,
	}, nil
}

func (s *Server) Start() error {
	slog.Info("Starting http server", "address", s.server.Addr)

	ln, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			ln.Close()
		}
	}()

	go s.server.Serve(ln)

	err = s.ravel.Start()
	if err != nil {
		slog.Error("Failed to start ravel", "error", err)
		return err
	}

	return nil
}

func (s *Server) Run(runCtx context.Context) {
	<-runCtx.Done()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Shutting down http server")
	s.server.Shutdown(ctxTimeout)

	err := s.ravel.Stop()
	if err != nil {
		slog.Error("Failed to stop ravel", "error", err)
	}
}

func getHumaConfig() huma.Config {
	return huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   "Ravel API",
				Version: "1.0.0",
			},
		},
		OpenAPIPath: "/openapi",
		DocsPath:    "/docs",
		Formats: map[string]huma.Format{
			"application/json": huma.DefaultJSONFormat,
			"json":             huma.DefaultJSONFormat,
		},
		DefaultFormat: "application/json",
	}
}
