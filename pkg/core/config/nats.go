package config

import "errors"

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
