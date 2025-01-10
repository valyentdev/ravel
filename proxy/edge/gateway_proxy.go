package edge

import (
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/proxy/httpproxy"
)

type edgeRCTX struct {
	instanceId string
	targetPort int
}

type gatewayProxyService struct {
	backends            *instanceBackends
	gateways            *gateways
	defaultDomainSuffix string
}

var _ interface {
	httpproxy.HTTPProxyService[edgeRCTX]
	httpproxy.WithRewrite[edgeRCTX]
} = (*gatewayProxyService)(nil)

func newGatewayProxyService(domainSuffix string, beStore *instanceBackends, corro *corroclient.CorroClient) *gatewayProxyService {
	gateways := newGateways(corro)
	gateways.Start()
	return &gatewayProxyService{
		backends:            beStore,
		gateways:            gateways,
		defaultDomainSuffix: domainSuffix,
	}
}

// GetUpstream implements ravelProxy.
func (p *gatewayProxyService) GetUpstream(r *http.Request, rctx *edgeRCTX) *url.URL {
	host := r.Host
	if strings.HasSuffix(host, p.defaultDomainSuffix) {
		host = strings.TrimSuffix(host, p.defaultDomainSuffix)
	} else {
		return nil
	}

	gw, ok := p.gateways.GetGateway(host)
	if !ok {
		slog.Debug("gateway not found", "host", host)
		return nil
	}

	machineId, ok := gw.Pick()
	if !ok {
		slog.Debug("no instances found", "gateway", gw)
		return nil
	}

	backend, ok := p.backends.getBackend(machineId)
	if !ok {
		slog.Debug("backend not found", "instanceId", backend.Id)
		return nil
	}

	rctx.instanceId = backend.Id
	rctx.targetPort = gw.TargetPort
	return backend.Url()
}

// NewRCTX implements ravelProxy.
func (p *gatewayProxyService) NewRCTX() *edgeRCTX {
	return &edgeRCTX{}
}

// Rewrite implements ravelProxy.
func (p *gatewayProxyService) Rewrite(pr *httpproxy.ProxyRequest, rctx *edgeRCTX) {
	pr.Out.Header.Set("Ravel-Instance-Port", strconv.Itoa(rctx.targetPort))
	pr.Out.Header.Set("Ravel-Instance-Id", rctx.instanceId)
	pr.Out.Host = pr.In.Host
}

type Gateway struct {
	api.Gateway
	Machines []string
}

func (gw *Gateway) Pick() (string, bool) {
	length := len(gw.Machines)
	if length == 0 {
		return "", false
	}

	idx := rand.Intn(length)

	return gw.Machines[idx], true
}

func (p *gatewayProxyService) Start() {
	p.gateways.Start()
}
