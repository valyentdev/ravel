package server

import (
	"context"
	"log/slog"
	"net/http"

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

func (s *Server) Serve() error {
	go s.ravel.ListenInstanceEvents()

	slog.Info("Starting http server", "address", s.server.Addr)

	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
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
