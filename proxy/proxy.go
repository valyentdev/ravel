package proxy

import (
	"errors"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/valyentdev/ravel/core/config"
)

type Mode string

const (
	Edge  Mode = "edge"
	Local Mode = "local"
)

type TLS struct {
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
}

type Config struct {
	Corrosion config.CorrosionConfig `toml:"corrosion"`
	Edge      EdgeConfig             `toml:"edge"`
	Local     LocalConfig            `toml:"local"`
}

type EdgeConfig struct {
	Address       string `toml:"address"`
	DefaultDomain string `toml:"default_domain"`
	TLS           TLS    `toml:"tls"`
}

type LocalConfig struct {
	Address string `toml:"address"`
	NodeId  string `toml:"node_id"`
}

func ReadConfigFile(path string) (*Config, error) {
	var cfg Config
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decoder := toml.NewDecoder(strings.NewReader(string(bytes)))
	decoder = decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		if sme, ok := err.(*toml.StrictMissingError); ok {
			return nil, errors.New(sme.String())
		}

		return nil, err
	}

	return &cfg, nil
}