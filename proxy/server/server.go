package server

import "net/http"

func NewServer(handler http.HandlerFunc, addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
