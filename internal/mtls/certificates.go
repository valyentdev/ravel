package mtls

import (
	"log/slog"
	"time"

	"github.com/cloudflare/cfssl/certinfo"
	"github.com/cloudflare/cfssl/cli/genkey"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/cloudflare/cfssl/signer/local"
)

type CertType = string

const (
	AgentCert  CertType = "agent"  // for agent, used for server auth and client auth
	ServerCert CertType = "server" // for api server auth, give authority on agents
	ClientCert CertType = "client" // for client (SDK, CLI...) to talk to api server
)

func certName(name, region string, certType CertType) string {
	return name + "." + region + "." + certType + ".ravel"
}

func NewCSR(name string) *csr.CertificateRequest {
	return &csr.CertificateRequest{
		CN: name,
		CA: csr.New().CA,
	}
}

func getAgentSigningPolicy(expiry time.Duration) *config.Signing {
	return &config.Signing{
		Default: &config.SigningProfile{
			Usage:  []string{"server auth", "client auth"},
			Expiry: expiry,
		},
	}
}

func getServerSigningPolicy(expiry time.Duration) *config.Signing {
	return &config.Signing{
		Default: &config.SigningProfile{
			Usage:  []string{"server auth", "client auth"},
			Expiry: expiry,
		},
	}
}

func getClientSigningPolicy(expiry time.Duration) *config.Signing {
	return &config.Signing{
		Default: &config.SigningProfile{
			Usage:  []string{"client auth"},
			Expiry: expiry,
		},
	}
}

func GenerateCA(cn string, expiry string) (certPem []byte, keyPem []byte, err error) {
	certPem, _, keyPem, err = initca.New(&csr.CertificateRequest{
		CN: cn,
		CA: &csr.CAConfig{
			Expiry: expiry,
		},
	})

	if err != nil {
		return
	}

	return
}

func generateCertPEM(caFile, caKeyFile string, config *config.Signing, req *csr.CertificateRequest) (certPem []byte, keyPem []byte, err error) {
	s, err := local.NewSignerFromFile(caFile, caKeyFile, config)
	if err != nil {
		slog.Error("Failed to create signer", "error", err)
		return nil, nil, err
	}

	generator := csr.Generator{Validator: genkey.Validator}

	csr, keyPem, err := generator.ProcessRequest(req)
	if err != nil {
		slog.Error("Failed to generate CSR", "error", err)
		return nil, nil, err
	}

	certPem, err = s.Sign(signer.SignRequest{
		Request: string(csr),
	})
	if err != nil {
		return
	}

	return
}
func GenerateServerCert(caFile, caKeyFile string, name, region string, expiry time.Duration, hosts []string) (certPem []byte, keyPem []byte, err error) {
	hosts = append([]string{"127.0.0.1"}, hosts...)
	csr := &csr.CertificateRequest{
		CN:    certName(name, region, ServerCert),
		Hosts: hosts,
	}

	return generateCertPEM(caFile, caKeyFile, getServerSigningPolicy(expiry), csr)
}

func GenerateAgentCert(caFile, caKeyFile string, name, region string, expiry time.Duration, hosts []string) (certPem []byte, keyPem []byte, err error) {
	hosts = append([]string{"127.0.0.1"}, hosts...)
	csr := &csr.CertificateRequest{
		CN:    certName(name, region, AgentCert),
		Hosts: hosts,
	}
	return generateCertPEM(caFile, caKeyFile, getAgentSigningPolicy(expiry), csr)
}

func GenerateClientCert(caFile, caKeyFile string, name, region string, expiry time.Duration) (certPem []byte, keyPem []byte, err error) {
	csr := &csr.CertificateRequest{
		CN: certName(name, region, ClientCert),
	}
	return generateCertPEM(caFile, caKeyFile, getClientSigningPolicy(expiry), csr)
}

func CertInfos(certPem []byte) (*certinfo.Certificate, error) {
	certificate, err := certinfo.ParseCertificatePEM(certPem)
	if err != nil {
		return nil, err
	}

	return certificate, nil
}

func RenewCA(caFile, caKeyFile string) (caPem []byte, err error) {
	return initca.RenewFromPEM(caFile, caKeyFile)
}
