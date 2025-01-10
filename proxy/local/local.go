package local

import (
	"net/http"

	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/httpproxy"
	"github.com/valyentdev/ravel/proxy/server"
)

func NewLocalProxyServer(config *proxy.Config) *server.Server {
	proxyService := newProxyService(config)

	proxyService.Start()

	proxy := httpproxy.NewProxy(proxyService, nil)

	server := server.NewServer()

	server.Register(&http.Server{
		Addr:    config.Local.Address,
		Handler: proxy,
	})

	return server
}
