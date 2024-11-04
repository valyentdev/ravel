package local

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/valyentdev/ravel/pkg/proxy/reverse"
)

func rewriteNoOp(pr *httputil.ProxyRequest) {
}

type Backend struct {
	proxy *httputil.ReverseProxy
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.proxy.ServeHTTP(w, r)
}

func newBackend(i Instance) *Backend {
	return &Backend{
		proxy: reverse.NewReverseProxy(&url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", i.Ip, i.Port),
		}, rewriteNoOp),
	}
}
