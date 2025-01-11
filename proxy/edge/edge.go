package edge

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/httpproxy"
	"github.com/valyentdev/ravel/proxy/server"
)

type edgeProxy struct {
	gws           httpproxy.ProxyHandler
	machines      httpproxy.ProxyHandler
	gatewayDomain string
	initdDomain   string
}

func loadCertificates(config *proxy.Config) (*certStore, error) {
	gwCert, err := tls.LoadX509KeyPair(config.Edge.TLS.CertFile, config.Edge.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	machinesCert, err := tls.LoadX509KeyPair(config.Edge.MachineGateways.TLS.CertFile, config.Edge.MachineGateways.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	return newCertStore(config.Edge.DefaultDomain, config.Edge.MachineGateways.Domain, &gwCert, &machinesCert), nil
}

func (p *edgeProxy) serveHTTPS(w http.ResponseWriter, r *http.Request) {
	host := httpproxy.StripHostPort(r.Host)

	// We'll need to check the SNI later when we'll support custom domains
	// for now, we'll just check the host as all domains are subdomains of the
	// defaults domains
	// sni := r.TLS.ServerNacme
	// if sni != host {
	// 	slog.Warn("SNI does not match host", "sni", sni, "host", host)
	// 	httpproxy.AnswerErrorStatus(w, r, http.StatusBadGateway)
	// 	return
	// }

	if isSubDomain(p.gatewayDomain, host) {
		p.gws.ServeHTTP(w, r)
		return
	}

	if isSubDomain(p.initdDomain, host) {
		p.machines.ServeHTTP(w, r)
		return
	}

	httpproxy.AnswerErrorStatus(w, r, http.StatusBadGateway)
}

func (p *edgeProxy) serveHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header()["Content-Type"] = nil

	newUrl := url.URL{
		Scheme:   "https",
		Host:     r.Host,
		RawQuery: r.URL.RawQuery,
		Path:     r.URL.EscapedPath(),
	}
	http.Redirect(w, r, newUrl.String(), http.StatusMovedPermanently)
}

func NewEdgeProxyServer(config *proxy.Config) (*server.Server, error) {
	corro := corroclient.NewCorroClient(config.Corrosion.Config())
	backends := newInstanceBackends(corro)
	backends.Start()

	gatewayDomainSuffix := "." + config.Edge.DefaultDomain
	machinesGwDomainSuffix := "." + config.Edge.MachineGateways.Domain

	gwService := newGatewayProxyService(gatewayDomainSuffix, backends, corro)
	gwProxy := httpproxy.NewProxy(gwService, nil)

	var authorizer Authorizer

	if config.Edge.MachineGateways.InitdAuthz != nil {
		authorizer = newValyentAuthorizer(config.Edge.MachineGateways.InitdAuthz.Endpoint)
	} else {
		slog.Warn("[DANGER] No authorization is configured for initd proxy")
		authorizer = &noAuthAuthorizer{}
	}

	initdProxyService := newInitdProxyService(backends, machinesGwDomainSuffix, authorizer)
	initdProxy := httpproxy.NewProxy(initdProxyService, nil)

	proxy := &edgeProxy{
		gws:           gwProxy,
		machines:      initdProxy,
		gatewayDomain: config.Edge.DefaultDomain,
		initdDomain:   config.Edge.MachineGateways.Domain,
	}

	certStore, err := loadCertificates(config)
	if err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:           config.Edge.HttpAddr,
		Handler:        http.HandlerFunc(proxy.serveHTTP),
		MaxHeaderBytes: 16 * 1024, // 16KB
	}

	httpsServer := &http.Server{
		Addr:      config.Edge.HttpsAddr,
		Handler:   http.HandlerFunc(proxy.serveHTTPS),
		TLSConfig: &tls.Config{GetCertificate: certStore.GetCertificate},

		// these fields will be tuned later and configurable
		MaxHeaderBytes: 16 * 1024, // 16KB
		IdleTimeout:    60 * time.Second,
		ReadTimeout:    10 * time.Minute,
		WriteTimeout:   10 * time.Minute,
	}

	server := server.NewServer()

	server.Register(httpServer)
	server.Register(httpsServer)

	return server, nil
}
