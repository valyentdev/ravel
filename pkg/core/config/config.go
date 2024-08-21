package config

import (
	"encoding/json"
	"os"

	"github.com/valyentdev/corroclient"
)

const LOGS_DIRECTORY = "/var/log/ravel"
const DAEMON_DB_PATH = "/var/lib/ravel/daemon.db"

type RavelConfig struct {
	RavelApi         ApiConfig                            `json:"ravel_api"`
	NodeId           string                               `json:"node_id"`
	Agent            AgentConfig                          `json:"agent"`
	Corrosion        corroclient.Config                   `json:"corrosion"`
	PostgresURL      string                               `json:"postgres_url"`
	MachineTemplates map[string]MachineResourcesTemplates `json:"machine_templates"`
	Nats             NatsConfig                           `json:"nats"`
}

type ApiConfig struct {
	Address string `json:"address"`
}

func ReadFile(path string) (RavelConfig, error) {
	var config RavelConfig

	bytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
