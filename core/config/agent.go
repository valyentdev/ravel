package config

import "github.com/alexisbouchez/ravel/api"

type AgentConfig struct {
	NodeId    string          `json:"node_id" toml:"node_id"`
	Region    string          `json:"region" toml:"region"`
	Address   string          `json:"address" toml:"address"`
	Port      int             `json:"port" toml:"port"`
	Resources api.Resources   `json:"resources" toml:"resources"`
	TLS       *TLSConfig      `json:"tls" toml:"tls"`
	BuildKit  *BuildKitConfig `json:"buildkit" toml:"buildkit"`
}

// BuildKitConfig holds configuration for the BuildKit image builder
type BuildKitConfig struct {
	Enabled             bool   `json:"enabled" toml:"enabled"`
	Socket              string `json:"socket" toml:"socket"`
	MaxConcurrentBuilds int    `json:"max_concurrent_builds" toml:"max_concurrent_builds"`
}

// DefaultBuildKitConfig returns the default BuildKit configuration
func DefaultBuildKitConfig() *BuildKitConfig {
	return &BuildKitConfig{
		Enabled:             false,
		Socket:              "unix:///run/buildkit/buildkitd.sock",
		MaxConcurrentBuilds: 2,
	}
}
