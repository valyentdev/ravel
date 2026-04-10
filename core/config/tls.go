package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

type TLSConfig struct {
	CaFile           string `json:"ca_file" toml:"ca_file"`
	CertFile         string `json:"cert_file" toml:"cert_file"`
	KeyFile          string `json:"key_file" toml:"key_file"`
	SkipVerifyServer bool   `json:"skip_verify_server" toml:"skip_verify_server"`
	SkipVerifyClient bool   `json:"skip_verify_client" toml:"skip_verify_client"`
}

func (c *TLSConfig) LoadCA() (*x509.CertPool, error) {
	bytes, err := os.ReadFile(c.CaFile)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()

	if !pool.AppendCertsFromPEM(bytes) {
		return nil, errors.New("failed to append CA certificate")
	}

	return pool, nil
}

func (c *TLSConfig) LoadCert() (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return tls.Certificate{}, err
	}

	return cert, nil
}
