package edge

import (
	"crypto/tls"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/initd"
	"github.com/valyentdev/ravel/proxy"
	"github.com/valyentdev/ravel/proxy/httpproxy"
	"github.com/valyentdev/ravel/proxy/server"
)

type edgeProxy struct {
	gws           httpproxy.ProxyHandler
	initd         httpproxy.ProxyHandler
	gatewayDomain string
	initdDomain   string
}

func loadCertificates(config *proxy.Config) (*certStore, error) {
	gwCert, err := tls.LoadX509KeyPair(config.Edge.TLS.CertFile, config.Edge.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	initdCert, err := tls.LoadX509KeyPair(config.Edge.Initd.TLS.CertFile, config.Edge.Initd.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	return newCertStore(config.Edge.DefaultDomain, config.Edge.Initd.Domain, &gwCert, &initdCert), nil
}

func (p *edgeProxy) serveHTTPS(w http.ResponseWriter, r *http.Request) {
	host := httpproxy.StripHostPort(r.Host)
	sni := r.TLS.ServerName
	if sni != host {
		httpproxy.AnswerErrorStatus(w, r, http.StatusBadGateway)
		return
	}

	if isSubDomain(p.gatewayDomain, host) {
		slog.Debug("proxying to a gateway")
		p.gws.ServeHTTP(w, r)
		return
	}

	if isSubDomain(p.initdDomain, host) {
		slog.Debug("proxying to initd")
		p.initd.ServeHTTP(w, r)
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
	initdDomainSuffix := "." + config.Edge.Initd.Domain

	gwService := newGatewayProxyService(gatewayDomainSuffix, backends, corro)
	gwProxy := httpproxy.NewProxy(gwService, nil)

	var authorizer Authorizer

	if config.Edge.Initd.Authz != nil {
		authorizer = newValyentAuthorizer(config.Edge.Initd.Authz.Endpoint)
	} else {
		slog.Warn("[DANGER] No authorization is configured for initd proxy")
		authorizer = &noAuthAuthorizer{}
	}

	initdProxyService := newInitdProxyService(backends, initdDomainSuffix, initd.InitdPort, authorizer)
	initdProxy := httpproxy.NewProxy(initdProxyService, nil)

	proxy := &edgeProxy{
		gws:           gwProxy,
		initd:         initdProxy,
		gatewayDomain: config.Edge.DefaultDomain,
		initdDomain:   config.Edge.Initd.Domain,
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
