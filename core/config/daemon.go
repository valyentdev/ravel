package config

type DaemonConfig struct {
	Socket       string         `json:"socket" toml:"socket"`
	DatabasePath string         `json:"database_path" toml:"database_path"`
	Agent        *AgentConfig   `json:"agent" toml:"agent"`
	Runtime      *RuntimeConfig `json:"runtime" toml:"runtime"`
}
