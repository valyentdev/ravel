package edge

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyentdev/ravel/proxy/httpproxy"
)

type initdRCTX struct {
	instanceId string
	namespace  string
}
type initdProxyService struct {
	authorizer   Authorizer
	backends     *instanceBackends
	domainSuffix string
	port         string
}

func newInitdProxyService(backends *instanceBackends, domainSuffix string, port int, authorizer Authorizer) *initdProxyService {
	return &initdProxyService{
		authorizer:   authorizer,
		backends:     backends,
		domainSuffix: domainSuffix,
		port:         strconv.Itoa(port),
	}
}

var _ interface {
	httpproxy.HTTPProxyService[initdRCTX]
	httpproxy.WithRewrite[initdRCTX]
	httpproxy.WithFilter[initdRCTX]
} = (*initdProxyService)(nil)

func (p *initdProxyService) NewRCTX() *initdRCTX {
	return &initdRCTX{}
}
func (p *initdProxyService) GetUpstream(r *http.Request, rctx *initdRCTX) *url.URL {
	slog.Debug("getting upstream initd")
	host := r.Host
	if !strings.HasSuffix(host, p.domainSuffix) {
		slog.Debug("host does not have suffix", "host", host, "suffix", p.domainSuffix)
		return nil
	}

	host = strings.TrimSuffix(host, p.domainSuffix)

	parts := strings.SplitN(host, "-", 2)
	if len(parts) != 2 {
		slog.Debug("host does not have two parts", "host", host)
		return nil
	}

	machineId := parts[0]

	instance, ok := p.backends.getBackend(machineId)
	if !ok {
		slog.Debug("instance not found", "machineId", machineId)
		return nil
	}

	if instance.Namespace != parts[1] {
		slog.Debug("namespace does not match", "instance", instance, "namespace", parts[1])
		return nil
	}

	rctx.instanceId = instance.Id
	rctx.namespace = instance.Namespace
	return instance.Url()
}

func (p *initdProxyService) Rewrite(r *httpproxy.ProxyRequest, rctx *initdRCTX) {
	r.Out.Header.Set("Ravel-Instance-Id", rctx.instanceId)
	r.Out.Header.Set("Ravel-Instance-Port", p.port)
}

func (p *initdProxyService) Filter(w http.ResponseWriter, r *http.Request, rctx *initdRCTX) bool {
	return true
}
