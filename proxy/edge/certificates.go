package edge

import (
	"crypto/tls"
	"strings"
)

func newCertStore(gwDomain, initdDomain string, gw, initd *tls.Certificate) *certStore {
	return &certStore{
		gwDomain:    gwDomain,
		gw:          gw,
		initdDomain: initdDomain,
		initd:       initd,
	}
}

// dumb certStore, to be replaced when support for custom domains is added
type certStore struct {
	gwDomain    string
	gw          *tls.Certificate
	initdDomain string
	initd       *tls.Certificate
}

func (cs *certStore) GetCertificate(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
	domain := chi.ServerName
	// Cut the first part of the domain
	_, after, found := strings.Cut(domain, ".")
	if !found {
		return nil, nil // the go tls package take care to send errNoCert
	}

	switch after {
	case cs.gwDomain:
		return cs.gw, nil
	case cs.initdDomain:
		return cs.initd, nil
	}
	return nil, nil // the go tls package take care to send errNoCert

}
