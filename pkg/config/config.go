package config

import (
	"encoding/json"
	"os"

	"github.com/valyentdev/ravel/pkg/helper/corroclient"
)

const LOGS_DIRECTORY = "/var/log/ravel"
const DAEMON_DB_PATH = "/var/lib/ravel/daemon.db"

type RavelConfig struct {
	NodeId       string             `json:"node_id"`
	ClusteringKV string             `json:"clustering_kv"`
	Server       ServerConfig       `json:"server"`
	Agent        AgentConfig        `json:"agent"`
	Corrosion    corroclient.Config `json:"corrosion"`
	Nats         NatsConfig         `json:"nats"`
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
