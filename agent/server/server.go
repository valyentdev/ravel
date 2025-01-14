package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/valyentdev/ravel/api/errdefs"
	"github.com/valyentdev/ravel/core/cluster"
)

type AgentServer struct {
	server *http.Server
	agent  cluster.Agent
}

func (e *AgentServer) log(msg string, err error) {
	var rerr *errdefs.RavelError
	slog.Debug(msg, "error", err)
	if errors.As(err, &rerr) {
		if errdefs.IsUnknown(err) || errdefs.IsInternal(err) {
			slog.Error(msg, "error", err)
		}
	} else {
		slog.Error(msg, "error", err)
	}
}

func (s *AgentServer) Serve(listener net.Listener) {
	s.server.Serve(listener)
}

func (s *AgentServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewAgentServer(agent cluster.Agent) *AgentServer {
	as := &AgentServer{agent: agent}

	mux := http.NewServeMux()

	as.registerEndpoints(mux)

	server := &http.Server{
		Handler: mux,
	}

	as.server = server

	return as
}
