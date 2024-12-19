package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/valyentdev/corroclient"
	"github.com/valyentdev/ravel/core/registry"
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
	Daemon     DaemonConfig              `json:"daemon" toml:"daemon"`
	Server     ServerConfig              `json:"server" toml:"server"`
	Corrosion  *CorrosionConfig          `json:"corrosion" toml:"corrosion"`
	Nats       *NatsConfig               `json:"nats" toml:"nats"`
	Registries registry.RegistriesConfig `json:"registries" toml:"registries"`
}

// never display data because it contains secrets
func fmtDecodeError(err *toml.DecodeError) error {
	line, column := err.Position()

	return fmt.Errorf("%s %s at %s", err.Error(), strings.Join(err.Key(), "."), fmt.Sprintf("line %d, column %d", line, column))

}

func joinErrors(errs ...error) error {
	var errStr string

	for _, err := range errs {
		errStr += err.Error() + "\n"
	}
	return errors.New(errStr)
}

func buildTomlError(err error) error {
	smeErr, ok := err.(*toml.StrictMissingError)
	if ok {
		var errs []error
		for _, e := range smeErr.Errors {
			errs = append(errs, fmtDecodeError(&e))
		}
		return joinErrors(errs...)
	}

	decodeErr, ok := err.(*toml.DecodeError)
	if ok {
		return joinErrors(fmtDecodeError(decodeErr))
	}

	return fmt.Errorf("toml error: %w", err)
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
			tomlErr := err.(*toml.StrictMissingError)
			return config, buildTomlError(tomlErr)
		}
	} else {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}
