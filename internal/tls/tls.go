package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func generateCA() (*x509.Certificate, *ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "Corrosion Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 5 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func generateServerCert(caCert *x509.Certificate, caPrivateKey *ecdsa.PrivateKey, ipAddr string) (*x509.Certificate, string, error) {
	serverPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, "", err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ravel.local",
		},
		IPAddresses: []net.IP{net.ParseIP(ipAddr)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, "", err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return cert, string(certPEM), nil
}

func generateClientCert(caCert *x509.Certificate, caPrivateKey *ecdsa.PrivateKey) (*x509.Certificate, string, error) {
	clientPrivateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return nil, "", err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, "", err
	}

	certTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1 year validity
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, // Client authentication
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, &clientPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return nil, "", err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, "", err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// If you need the private key as PEM as well, uncomment the following lines:
	// privBytes, err := x509.MarshalECPrivateKey(clientPrivateKey)
	// if err != nil {
	// 	return nil, "", err
	// }
	// privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})

	return cert, string(certPEM), nil
}
