package config

import (
	"errors"

	"github.com/valyentdev/ravel/pkg/core"
)

type Resources struct {
	Cpus   int `json:"cpus" toml:"cpus"`     // in MHz
	Memory int `json:"memory" toml:"memory"` // in MB
}

func (r Resources) Resources() core.Resources {
	return core.Resources{
		Cpus:   r.Cpus,
		Memory: r.Memory,
	}
}

type AgentConfig struct {
	NodeId        string    `json:"node_id" toml:"node_id"`
	Region        string    `json:"region" toml:"region"`
	Address       string    `json:"address" toml:"address"`
	AgentPort     int       `json:"agent_port" toml:"agent_port"`
	HttpProxyPort int       `json:"http_proxy_port" toml:"http_proxy_port"`
	InitBinary    string    `json:"init_binary" toml:"init_binary"`
	LinuxKernel   string    `json:"linux_kernel" toml:"linux_kernel"`
	DatabasePath  string    `json:"database_path" toml:"database_path"`
	Resources     Resources `json:"resources" toml:"resources"`
}

func (dc AgentConfig) Validate() error {
	if dc.InitBinary == "" {
		return errors.New("init_binary is required")
	}
	if dc.LinuxKernel == "" {
		return errors.New("linux_kernel is required")
	}

	if dc.DatabasePath == "" {
		return errors.New("database_path is required")
	}
	return nil

}
