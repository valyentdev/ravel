package edge

import (
	"log/slog"
	"math/rand"
	"net/http"
	"strings"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/proxy"
)

type RavelProxy struct {
	backends      *Backends
	gateways      *Gateways
	defaultDomain string
}

func NewRavelProxy(config *proxy.Config) *RavelProxy {
	corro := corroclient.NewCorroClient(config.Corrosion.Config())
	return &RavelProxy{
		backends:      newBackends(corro),
		gateways:      newGateways(corro),
		defaultDomain: "bigworld.app",
	}
}

type Gateway struct {
	api.Gateway
	Instances []string `json:"instances"`
}

func (gw *Gateway) Pick() (string, bool) {
	length := len(gw.Instances)
	if length == 0 {
		return "", false
	}

	idx := rand.Intn(length)

	return gw.Instances[idx], true
}

func (p *RavelProxy) Start() {
	p.backends.Start()
	p.gateways.Start()
}

func (p *RavelProxy) Handle(w http.ResponseWriter, r *http.Request) {
	slog.Debug("proxy request")
	host := r.Host
	suffix := "." + p.defaultDomain
	if strings.HasSuffix(host, suffix) {
		host = strings.TrimSuffix(host, suffix)
	} else {
		w.WriteHeader(http.StatusBadGateway)
		slog.Info("bad gateway")
		return
	}

	gw, ok := p.gateways.GetGateway(host)
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		slog.Info("bad gateway")
		return
	}

	instanceId, ok := gw.Pick()
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		slog.Info("bad gateway")
		return
	}

	backend, ok := p.backends.getBackend(instanceId)
	if !ok {
		w.WriteHeader(http.StatusBadGateway)
		slog.Info("bad gateway")
		return
	}

	slog.Debug("proxy to backend")
	backend.ServeHTTP(w, r)

}
