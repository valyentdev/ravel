package ravelctl

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	APIURL string `json:"api_url"`
	Token  string `json:"token"`
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ravel", "config.json")
}

func LoadConfig() (*Config, error) {
	return LoadConfigFrom(DefaultConfigPath())
}

func LoadConfigFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				APIURL: "http://localhost:3000",
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.APIURL == "" {
		cfg.APIURL = "http://localhost:3000"
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	return c.SaveTo(DefaultConfigPath())
}

func (c *Config) SaveTo(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}
