package edge

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/valyentdev/ravel/proxy/reverse"
)

type Instance struct {
	Id         string `json:"id"`
	Gatewayid  string `json:"gateway_id"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
	TargetPort int    `json:"target_port"`
}

type Backend struct {
	instance Instance
	proxy    *httputil.ReverseProxy
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("proxy request", "instance_id", b.instance.Id, "node", b.instance.Address, "port", b.instance.Port)
	b.proxy.ServeHTTP(w, r)
}

func newBackend(i Instance) *Backend {
	url, _ := url.Parse(fmt.Sprintf("http://%s:%d", i.Address, i.Port))
	return &Backend{
		instance: i,
		proxy: reverse.NewReverseProxy(url, func(pr *httputil.ProxyRequest) {
			slog.Debug("rewriting request", "instance_id", i.Id, "gateway_id", i.Gatewayid)
			pr.Out.Header.Set("Ravel-Instance-Id", i.Id)
			pr.Out.Header.Set("Ravel-Gateway-Id", i.Gatewayid)
		}),
	}
}
