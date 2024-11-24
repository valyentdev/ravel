package config

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
)

type NatsConfig struct {
	Url      string `json:"url" toml:"url"`
	CredFile string `json:"cred_file" toml:"cred_file"`
}

func (n NatsConfig) Validate() error {
	if n.Url == "" {
		return errors.New("nats.url is required")
	}
	if n.CredFile == "" {
		return errors.New("nats.cred_file is required")
	}
	return nil
}

func (n NatsConfig) Connect() (*nats.Conn, error) {
	slog.Info("Connecting to nats", "url", n.Url, "cred_file", n.CredFile)
	natsOptions := []nats.Option{}
	if n.CredFile != "" {
		natsOptions = append(natsOptions, nats.UserCredentials(n.CredFile, n.CredFile))
		natsOptions = append(natsOptions, nats.MaxReconnects(-1))
	}

	nc, err := nats.Connect(n.Url, natsOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	return nc, nil
}
