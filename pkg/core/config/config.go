package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/valyentdev/corroclient"
)

const LOGS_DIRECTORY = "/var/log/ravel"
const DAEMON_DB_PATH = "/var/lib/ravel/daemon.db"

type CorrosionConfig struct {
	URL    string `json:"url" toml:"url"`
	Bearer string `json:"bearer" toml:"bearer"`
}

func (cc CorrosionConfig) Config() corroclient.Config {
	return corroclient.Config{
		URL:    cc.URL,
		Bearer: cc.Bearer,
	}
}

type RavelConfig struct {
	Server    ServerConfig    `json:"server" toml:"server"`
	Agent     AgentConfig     `json:"agent" toml:"agent"`
	Corrosion CorrosionConfig `json:"corrosion" toml:"corrosion"`
	Nats      NatsConfig      `json:"nats" toml:"nats"`
}

func ReadFile(path string) (RavelConfig, error) {
	var config RavelConfig

	bytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if strings.HasSuffix(path, ".toml") {
		decoder := toml.NewDecoder(strings.NewReader(string(bytes)))
		decoder = decoder.DisallowUnknownFields()
		err = decoder.Decode(&config)
		if err != nil {
			if sme, ok := err.(*toml.StrictMissingError); ok {
				return config, errors.New(sme.String())
			}
		}
	} else {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}
