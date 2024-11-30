package config

import "github.com/valyentdev/ravel/api"

type AgentConfig struct {
	NodeId        string        `json:"node_id" toml:"node_id"`
	Region        string        `json:"region" toml:"region"`
	Address       string        `json:"address" toml:"address"`
	Port          int           `json:"port" toml:"port"`
	HttpProxyPort int           `json:"http_proxy_port" toml:"http_proxy_port"`
	Resources     api.Resources `json:"resources" toml:"resources"`
	TLS           *TLSConfig    `json:"tls" toml:"tls"`
}
