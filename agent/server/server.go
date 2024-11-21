package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/valyentdev/ravel/agent/structs"
	"github.com/valyentdev/ravel/core/errdefs"
)

type AgentServer struct {
	server *http.Server
	agent  structs.Agent
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

func (s *AgentServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *AgentServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewAgentServer(agent structs.Agent, address string) *AgentServer {
	as := &AgentServer{agent: agent}

	mux := http.NewServeMux()
	as.registerEndpoints(mux)

	server := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	as.server = server

	return as
}
