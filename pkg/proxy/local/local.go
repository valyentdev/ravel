package local

import (
	"log/slog"
	"net/http"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/pkg/proxy"
)

func NewProxy(config *proxy.Config) *Proxy {
	corro := corroclient.NewCorroClient(config.Corrosion.Config())
	return &Proxy{
		instances: newInstances(corro, config.Local.NodeId),
		corro:     corro,
	}
}

type Proxy struct {
	instances *instances
	corro     *corroclient.CorroClient
}

func (p *Proxy) Start() {
	p.instances.Start()
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	instanceId := r.Header.Get("Ravel-Instance-Id")
	gatewayId := r.Header.Get("Ravel-Gateway-Id")
	if instanceId == "" || gatewayId == "" {
		slog.Debug("bad gateway id or instance id")
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	slog.Debug("proxy request", "instance_id", instanceId, "gateway_id", gatewayId)

	instance, ok := p.instances.getInstance(gatewayId)
	if !ok {
		slog.Debug("instance not found", "instance_id", instanceId, "gateway_id", gatewayId)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	instance.ServeHTTP(w, r)
}
