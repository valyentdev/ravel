package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	servers []*http.Server
}

func NewServer() *Server {
	return &Server{
		servers: make([]*http.Server, 0),
	}
}

func (s *Server) Register(server *http.Server) {
	s.servers = append(s.servers, server)
}

func (s *Server) Start() error {
	listeners := make([]net.Listener, len(s.servers))
	for i, server := range s.servers {
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			return err
		}
		listeners[i] = listener
	}

	for i, server := range s.servers {
		go func(i int, server *http.Server) {
			var err error
			if server.TLSConfig != nil {
				err = server.ServeTLS(listeners[i], "", "")
			} else {
				err = server.Serve(listeners[i])
			}

			if err != http.ErrServerClosed {
				panic(err)
			}
		}(i, server)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) {
	wg := sync.WaitGroup{}

	wg.Add(len(s.servers))
	for _, server := range s.servers {
		go func(server *http.Server) {
			err := server.Shutdown(ctx)
			if err != nil {
				slog.Error("error shutting down server", "error", err)
			}

			wg.Done()
		}(server)
	}

	wg.Wait()
}
