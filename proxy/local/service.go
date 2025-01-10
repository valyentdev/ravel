package local

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/httpproxy"
)

func newProxyService(config *proxy.Config) *proxyService {
	corro := corroclient.NewCorroClient(config.Corrosion.Config())
	return &proxyService{
		instances: newInstances(corro, config.Local.NodeId),
		corro:     corro,
	}
}

type lpRCTX struct {
	instanceId string
	targetPort string
}

type localProxy interface {
	httpproxy.HTTPProxyService[lpRCTX]
	httpproxy.WithRewrite[lpRCTX]
}

type proxyService struct {
	instances *instances
	corro     *corroclient.CorroClient
}

// GetUpstream implements localProxy.
func (p *proxyService) GetUpstream(r *http.Request, rctx *lpRCTX) *url.URL {
	slog.Debug("getting upstream")
	instanceId := r.Header.Get("Ravel-Instance-Id")
	portStr := r.Header.Get("Ravel-Instance-Port")

	_, err := strconv.Atoi(portStr)
	if err != nil || instanceId == "" || portStr == "" {
		slog.Debug("bad instance id or port")
		return nil
	}

	instance, ok := p.instances.getInstance(instanceId)
	if !ok {
		slog.Debug("instance not found", "instanceId", instanceId)
		return nil
	}
	rctx.instanceId = instanceId
	rctx.targetPort = portStr
	slog.Debug("instance found", "instanceId", instanceId, "port", portStr)
	return instance.Url(portStr)
}

// NewRCTX implements localProxy.
func (p *proxyService) NewRCTX() *lpRCTX {
	return &lpRCTX{}
}

// Rewrite implements localProxy.
func (p *proxyService) Rewrite(pr *httpproxy.ProxyRequest, rctx *lpRCTX) {
	pr.Out.Header.Set("Ravel-Instance-Port", rctx.targetPort)
	pr.Out.Header.Set("Ravel-Instance-Id", rctx.instanceId)
	pr.Out.Host = pr.In.Host // keep the original host
}

var _ localProxy = (*proxyService)(nil)

func (p *proxyService) Start() {
	p.instances.Start()
}
