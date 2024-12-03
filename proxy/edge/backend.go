package edge

import (
	"fmt"
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
	b.proxy.ServeHTTP(w, r)
}

func newBackend(i Instance) *Backend {
	url, _ := url.Parse(fmt.Sprintf("http://%s:%d", i.Address, i.Port))
	return &Backend{
		instance: i,
		proxy: reverse.NewReverseProxy(url, func(pr *httputil.ProxyRequest) {
			pr.Out.Header.Set("Ravel-Instance-Id", i.Id)
			pr.Out.Header.Set("Ravel-Gateway-Id", i.Gatewayid)
		}),
	}
}
