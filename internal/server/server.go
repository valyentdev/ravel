package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/valyentdev/ravel/internal/server/endpoints"
	"github.com/valyentdev/ravel/internal/server/utils"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/config"
	"github.com/valyentdev/ravel/pkg/ravel"
)

type Server struct {
	ravel     *ravel.Ravel
	server    *http.Server
	validator *utils.Validator
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

	core.OverrideHumaErrorBuilder()

	mux := http.NewServeMux()

	humaConfig := utils.GetHumaConfig()
	api := humago.New(mux, humaConfig)
	e := endpoints.New(r)
	e.Register(api)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return &Server{
		ravel:     r,
		server:    server,
		validator: utils.NewValidator(),
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
