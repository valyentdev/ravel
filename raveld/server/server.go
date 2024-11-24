package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/errdefs"
)

type DaemonServer struct {
	server *http.Server
	daemon daemon.Daemon
}

func (e *DaemonServer) log(msg string, err error) {
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

func (s *DaemonServer) Serve(ln net.Listener) {
	s.server.Serve(ln)
}
func (s *DaemonServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewDaemonServer(daemon daemon.Daemon) *DaemonServer {
	as := &DaemonServer{daemon: daemon}

	mux := http.NewServeMux()
	as.registerEndpoints(mux)

	server := &http.Server{
		Handler: mux,
	}

	as.server = server

	return as
}
