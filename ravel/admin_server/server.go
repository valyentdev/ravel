package server

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"

	"github.com/valyentdev/ravel/core/config"
	"github.com/valyentdev/ravel/core/errdefs"
	"github.com/valyentdev/ravel/internal/mtls"
	"github.com/valyentdev/ravel/ravel"
	"github.com/valyentdev/ravel/ravel/server/endpoints"
)

const (
	DEFAULT_ADMIN_SERVER_ADDR = ":3001"
)

// AdminServer represents an administration HTTP API server.
type AdminServer struct {
	config *config.ServerAPIConfig
	bearer []byte
	ravel  *ravel.Ravel
	server *http.Server
}

// NewAdminServer creates a new instance of AdminServer.
func NewAdminServer(c config.RavelConfig) (*AdminServer, error) {
	r, err := ravel.New(c)
	if err != nil {
		return nil, err
	}

	address := c.Server.API.Address
	if address == "" {
		address = DEFAULT_ADMIN_SERVER_ADDR
	}

	errdefs.OverrideHumaErrorBuilder()

	mux := http.NewServeMux()

	humaConfig := getHumaConfig()
	api := humago.New(mux, humaConfig)
	e := endpoints.New(r)

	var bearer []byte
	if c.Server.API.BearerToken != "" {
		bearer = []byte(c.Server.API.BearerToken)
	}

	api.UseMiddleware(newAuthMiddleware(bearer))

	e.Register(api)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	return &AdminServer{
		ravel:  r,
		server: server,
		bearer: []byte(c.Server.API.BearerToken),
		config: &c.Server.API,
	}, nil
}

func (srv *AdminServer) listen() (net.Listener, error) {
	address := srv.config.Address
	if address == "" {
		address = DEFAULT_ADMIN_SERVER_ADDR
	}

	if srv.config.TLS == nil {
		return net.Listen("tcp", address)
	}

	cert, err := srv.config.TLS.LoadCert()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if !srv.config.TLS.SkipVerifyClient {
		ca, err := srv.config.TLS.LoadCA()
		if err != nil {
			return nil, err
		}

		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = ca
		tlsConfig.VerifyConnection = mtls.VerifyServerAPIConnection
	}

	return tls.Listen("tcp", address, tlsConfig)
}

func (srv *AdminServer) Start() error {
	slog.Info("Starting http server", "address", srv.server.Addr)

	ln, err := srv.listen()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			ln.Close()
		}
	}()

	go srv.server.Serve(ln)

	err = srv.ravel.Start()
	if err != nil {
		slog.Error("Failed to start ravel", "error", err)
		return err
	}

	return nil
}

func (s *AdminServer) Run(runCtx context.Context) {
	<-runCtx.Done()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Shutting down http server")
	s.server.Shutdown(ctxTimeout)
}

func getHumaConfig() huma.Config {
	return huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   "Ravel Administration API",
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
