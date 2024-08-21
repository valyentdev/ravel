package config

import (
	"errors"

	"github.com/valyentdev/ravel/pkg/core"
)

type Resources = core.Resources

type AgentConfig struct {
	Region        string    `json:"region"`
	Address       string    `json:"address"`
	InitBinary    string    `json:"init_binary"`
	LinuxKernel   string    `json:"linux_kernel"`
	LogsDirectory string    `json:"logs_directory"`
	DatabasePath  string    `json:"database_path"`
	Resources     Resources `json:"resources"`
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
