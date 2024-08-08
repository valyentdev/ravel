package main

import (
	"encoding/json"
	"os"

	"github.com/valyentdev/ravel/internal/vminit"
)

func DecodeConfig(path string) (vminit.Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return vminit.Config{}, err
	}

	var config vminit.Config

	if err := json.Unmarshal(contents, &config); err != nil {
		return vminit.Config{}, err
	}

	return config, nil
}
