package tls

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/cloudflare/cfssl/certinfo"
	"github.com/cloudflare/cfssl/log"
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/internal/mtls"
)

func NewTLSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tls",
		Short: "Ravel TLS certificates management for mTLS",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Level = log.LevelWarning
		},
	}

	cmd.AddCommand(newInitCACmd())
	cmd.AddCommand(newCreateCertCmd())
	cmd.AddCommand(newCertInfosCmd())

	return cmd
}

const year = 8760 * time.Hour

func newInitCACmd() *cobra.Command {
	log.Level = log.LevelError
	cn := "Ravel CA"
	expiry := 5 * year
	output := "."

	cmd := &cobra.Command{
		Use:   "ca-init",
		Short: "Generate a new CA certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			ca, key, err := mtls.GenerateCA(cn, expiry.String())
			if err != nil {
				return err
			}

			err = writeCertificate(output, "ravel-ca", ca, key, 0611, 0600)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&cn, "cn", cn, "Common Name for the CA")
	cmd.Flags().DurationVar(&expiry, "expiry", expiry, "Expiry duration for the CA")
	cmd.Flags().StringVarP(&output, "output", "o", output, "Folder to save the CA certificate and key")

	return cmd
}

func createCert(
	certType mtls.CertType,
	caFile, caKeyFile string,
	name, region string,
	expiry time.Duration,
	hosts []string,
) (cert []byte, key []byte, err error) {
	switch certType {
	case mtls.AgentCert:
		cert, key, err = mtls.GenerateAgentCert(caFile, caKeyFile, name, region, expiry, hosts)
	case mtls.ServerCert:
		cert, key, err = mtls.GenerateServerCert(caFile, caKeyFile, name, region, expiry, hosts)
	case mtls.ClientCert:
		cert, key, err = mtls.GenerateClientCert(caFile, caKeyFile, name, region, expiry)
	default:
		err = fmt.Errorf("unknown certificate type %s", certType)
	}
	return
}

func newCreateCertCmd() *cobra.Command {
	var (
		output    string
		caFile    string
		caKeyFile string
		name      string
		expiry    time.Duration
		region    string
		hosts     []string

		server bool
		client bool
		agent  bool
	)

	cmd := &cobra.Command{
		Use:   "cert-create",
		Short: "Generate a new certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			var certType mtls.CertType
			if server && !client && !agent {
				certType = mtls.ServerCert
			} else if client && !server && !agent {
				certType = mtls.ClientCert
			} else if agent && !server && !client {
				certType = mtls.AgentCert
			} else {
				return errors.New("you must specify exactly one of --server, --client, or --agent")
			}

			cert, key, err := createCert(certType, caFile, caKeyFile, name, region, expiry, hosts)
			if err != nil {
				return err
			}

			err = writeCertificate(output, fmt.Sprintf("%s-%s", name, certType), cert, key, 0644, 0600)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&caFile, "ca", "", "CA certificate file")
	cmd.Flags().StringVar(&caKeyFile, "key", "", "CA key file")
	cmd.MarkFlagRequired("ca")
	cmd.MarkFlagRequired("key")
	cmd.Flags().StringVarP(&output, "output", "o", ".", "Folder to save the certificate and key")
	cmd.Flags().BoolVar(&server, "server", false, "Generate a server certificate")
	cmd.Flags().BoolVar(&agent, "agent", false, "Generate an agent certificate")
	cmd.Flags().BoolVar(&client, "client", false, "Generate a client certificate")

	cmd.Flags().StringVarP(&name, "name", "n", "ravel-1", "Name of the node you generate the certificate for")
	cmd.Flags().DurationVar(&expiry, "expiry", 1*year, "Expiry duration for the certificate")
	cmd.Flags().StringVar(&region, "region", "global", "Region for the certificate")
	cmd.Flags().StringSliceVar(&hosts, "hosts", nil, "Hosts for the certificate")

	return cmd
}

func newCertInfosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert-infos <cert>",
		Short: "Get certificate informations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("you must specify a certificate file")
			}

			cert := args[0]

			certPem, err := os.ReadFile(cert)
			if err != nil {
				return err
			}

			infos, err := mtls.CertInfos(certPem)
			if err != nil {
				return err
			}

			printCertificateInfos(infos)

			return nil
		},
	}

	return cmd
}

func newCARenewCmd() *cobra.Command {
	var (
		output    string
		caFile    string
		caKeyFile string
	)

	cmd := &cobra.Command{
		Use:   "ca-renew",
		Short: "Renew a CA certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			certPem, err := mtls.RenewCA(caFile, caKeyFile)
			if err != nil {
				return err
			}

			err = writeIfNotExists(output, "ravel-ca", certPem, 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&caFile, "ca", "", "CA certificate file")
	cmd.Flags().StringVar(&caKeyFile, "key", "", "CA key file")
	cmd.MarkFlagRequired("ca")
	cmd.MarkFlagRequired("key")
	cmd.Flags().StringVarP(&output, "output", "o", ".", "Folder to save the CA certificate and key")

	return cmd
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil

}

func writeIfNotExists(output, name string, data []byte, keyFileMode fs.FileMode) error {
	err := os.MkdirAll(output, keyFileMode)
	if err != nil {
		return err
	}

	file := path.Join(output, name)

	fileExists, err := fileExists(file)
	if err != nil {
		return err
	}

	if fileExists {
		return fmt.Errorf("file %s already exists", file)
	}

	err = os.WriteFile(file, data, keyFileMode)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %s to %s\n", name, file)

	return nil
}

func writeCertificate(output, name string, cert []byte, key []byte, certFileMode fs.FileMode, keyFileMode fs.FileMode) error {
	err := os.MkdirAll(output, certFileMode)
	if err != nil {
		return err
	}

	certFile := fmt.Sprintf("%s-cert.pem", name)
	keyFile := fmt.Sprintf("%s-key.pem", name)

	certFileExists, err := fileExists(path.Join(output, certFile))
	if err != nil {
		return err
	}

	if certFileExists {
		return fmt.Errorf("certificate file %s already exists", certFile)
	}

	keyFileExists, err := fileExists(path.Join(output, keyFile))
	if err != nil {
		return err
	}

	if keyFileExists {
		return fmt.Errorf("key file %s already exists", keyFile)
	}

	err = os.WriteFile(path.Join(output, certFile), cert, certFileMode)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %s certificate to %s\n", name, path.Join(output, certFile))

	err = os.WriteFile(path.Join(output, keyFile), key, keyFileMode)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote %s key to %s\n", name, path.Join(output, certFile))

	return nil
}

func printCertificateInfos(infos *certinfo.Certificate) {
	fmt.Println("Subject:", infos.Subject.CommonName)
	fmt.Println("Issuer:", infos.Issuer.CommonName)
	fmt.Println("NotBefore:", infos.NotBefore)
	fmt.Println("NotAfter:", infos.NotAfter)
	fmt.Println("SerialNumber:", infos.SerialNumber)
	fmt.Println("SignatureAlgorithm:", infos.SignatureAlgorithm)
}
