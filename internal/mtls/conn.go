package mtls

import (
	"crypto/tls"
	"errors"
	"strings"
)

var ErrInvalidClientCert = errors.New("invalid certificate")

func VerifyAgentConnection(cs tls.ConnectionState) error {
	for _, cert := range cs.PeerCertificates {
		// <name>.<region>.<certType>.ravel
		sn := strings.Split(cert.Subject.CommonName, ".")
		if len(sn) != 4 {
			return ErrInvalidClientCert
		}

		if sn[2] != "server" {
			return ErrInvalidClientCert
		}
	}
	return nil
}
