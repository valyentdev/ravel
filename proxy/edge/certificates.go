package edge

import (
	"crypto/tls"
	"strings"
)

func newCertStore(gwDomain, machinesGwDomain string, gw, machineGw *tls.Certificate) *certStore {
	return &certStore{
		gwDomain:         gwDomain,
		gw:               gw,
		machinesGwDomain: machinesGwDomain,
		machinesGw:       machineGw,
	}
}

// dumb certStore, to be replaced when support for custom domains is added
type certStore struct {
	gwDomain         string
	gw               *tls.Certificate
	machinesGwDomain string
	machinesGw       *tls.Certificate
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
	case cs.machinesGwDomain:
		return cs.machinesGw, nil
	}
	return nil, nil // the go tls package take care to send errNoCert

}
