package core

import "time"

type (
	RegistryAuthConfig struct {
		ServerAddress string `json:"server_address"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		Auth          string `json:"auth"`
		IdentityToken string `json:"identitytoken"`
	}

	GuestConfig struct {
		CpuKind  string `json:"cpu_kind"`
		MemoryMB int    `json:"memory_mb" minimum:"1"`
		Cpus     int    `json:"cpus" minimum:"1"`
	}

	Workload struct {
		Runtime       string              `json:"type"`
		Image         string              `json:"image"`
		RestartPolicy RestartPolicyConfig `json:"restart_policy,omitempty"`
		Env           []string            `json:"env,omitempty"`
		Init          InitConfig          `json:"init,omitempty"`
	}

	InitConfig struct {
		Cmd        []string `json:"cmd,omitempty"`
		Entrypoint []string `json:"entrypoint,omitempty"`
		User       string   `json:"user,omitempty"`
	}

	RestartPolicyConfig struct {
		MaxRetries int           `json:"max_retries,omitempty"`
		Policy     RestartPolicy `json:"policy,omitempty"`
	}
)

type RestartPolicy string

const (
	RestartPolicyAlways    RestartPolicy = "always"
	RestartPolicyOnFailure RestartPolicy = "on-failure"
	RestartPolicyNever     RestartPolicy = "never"
)

type StopConfig struct {
	Timeout *int    `json:"timeout,omitempty"`
	Signal  *string `json:"signal,omitempty"`
}

func (s *StopConfig) GetTimeout() time.Duration {
	if s.Timeout == nil {
		return 10 * time.Second
	}
	return time.Duration(*s.Timeout) * time.Second
}

func (s *StopConfig) GetSignal() string {
	if s.Signal == nil {
		return "SIGTERM"
	}
	return *s.Signal
}
