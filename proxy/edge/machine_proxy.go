package edge

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/proxy/httpproxy"
)

type machineRCTX struct {
	instanceId string
	namespace  string
	needAuth   bool
	port       string
}
type machineProxyService struct {
	authorizer   Authorizer
	backends     *instanceBackends
	domainSuffix string
}

func newInitdProxyService(backends *instanceBackends, domainSuffix string, authorizer Authorizer) *machineProxyService {
	return &machineProxyService{
		authorizer:   authorizer,
		backends:     backends,
		domainSuffix: domainSuffix,
	}
}

var _ interface {
	httpproxy.HTTPProxyService[machineRCTX]
	httpproxy.WithRewrite[machineRCTX]
	httpproxy.WithFilter[machineRCTX]
} = (*machineProxyService)(nil)

func (p *machineProxyService) NewRCTX() *machineRCTX {
	return &machineRCTX{}
}

func getPort(part string) (portStr string, isInitd bool, ok bool) {
	if part == "initd" {
		return initd.InitdPortStr, true, true
	}

	if part == initd.InitdPortStr {
		return initd.InitdPortStr, true, true
	}

	_, err := strconv.Atoi(part)
	if err != nil {
		slog.Debug("part is not a number", "part", part)
		return "", false, false
	}

	return part, false, true

}
func (p *machineProxyService) GetUpstream(r *http.Request, rctx *machineRCTX) *url.URL {
	host := r.Host
	host, _, found := strings.Cut(host, ".")
	if !found {
		slog.Debug("host does not have a suffix", "host", host)
		return nil
	}

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

	port, isInitd, ok := getPort(parts[1])
	if !ok {
		slog.Debug("port not found", "host", host)
		return nil
	}

	if !isInitd && !instance.EnableMachineGateway {
		slog.Debug("machine gateway not enabled", "instance", instance)
		return nil
	}

	rctx.port = port
	rctx.needAuth = isInitd
	rctx.instanceId = instance.Id
	rctx.namespace = instance.Namespace

	return instance.Url()
}

func (p *machineProxyService) Rewrite(r *httpproxy.ProxyRequest, rctx *machineRCTX) {
	r.Out.Header.Set("Ravel-Instance-Id", rctx.instanceId)
	r.Out.Header.Set("Ravel-Instance-Port", rctx.port)
}

func (p *machineProxyService) Filter(w http.ResponseWriter, r *http.Request, rctx *machineRCTX) bool {
	if rctx.needAuth {
		if !p.authorizer.Authorize(r, rctx.namespace) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			slog.Debug("unauthorized request", "rctx", rctx)
			return false
		}
	}

	return true
}
