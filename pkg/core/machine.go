package core

import (
	"time"

	"github.com/valyentdev/ravel/pkg/proto"
)

type Machine struct {
	Id         string        `json:"id"`
	Namespace  string        `json:"namespace"`
	FleetId    string        `json:"fleet"`
	InstanceId string        `json:"instance_id"`
	Region     string        `json:"region"`
	Config     MachineConfig `json:"config"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
	State      MachineStatus `json:"state"`
}

type MachineStatus string

const (
	MachineStatusCreated    MachineStatus = "created"
	MachineStatusPreparing  MachineStatus = "preparing"
	MachineStatusStarting   MachineStatus = "starting"
	MachineStatusRunning    MachineStatus = "running"
	MachineStatusStopping   MachineStatus = "stopping"
	MachineStatusStopped    MachineStatus = "stopped"
	MachineStatusDestroying MachineStatus = "destroying"
	MachineStatusDestroyed  MachineStatus = "destroyed"
)

type (
	RegistryAuthConfig struct {
		ServerAddress string `json:"server_address"`
		Username      string `json:"username"`
		Password      string `json:"password"`
		Auth          string `json:"auth"`
		IdentityToken string `json:"identitytoken"`
	}

	MachineConfig struct {
		Guest       GuestConfig `json:"guest"`
		Workload    Workload    `json:"workload"`
		StopConfig  StopConfig  `json:"stop_config,omitempty"`
		AutoDestroy bool        `json:"auto_destroy,omitempty"`
	}

	GuestConfig struct {
		MemoryMB int64 `json:"memory_mb"`
		VCpus    int64 `json:"vcpus"` // number of virtual CPUs (correspond to ch vcpus process threads)
		Cpus     int64 `json:"cpus"`  // in MHz
	}

	Workload struct {
		Kind          WorkloadKind        `json:"type"`
		Image         string              `json:"image"`
		RestartPolicy RestartPolicyConfig `json:"restart_policy,omitempty"`
		Env           []string            `json:"env,omitempty"`
		Init          InitConfig          `json:"init,omitempty"`
	}

	StopConfig struct {
		Timeout int    `json:"timeout"`
		Signal  string `json:"signal"`
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

type WorkloadKind string

const (
	WorkloadKindContainer WorkloadKind = "container"
)

func MachineConfigFromProto(m *proto.MachineConfig) MachineConfig {
	return MachineConfig{
		Guest: GuestConfig{
			Cpus:     m.Guest.Cpus,
			MemoryMB: m.Guest.MemoryMb,
		},
		Workload: Workload{
			Kind:  WorkloadKind(m.Workload.Kind),
			Image: m.Workload.Image,
			Env:   m.Workload.Env,
			Init: InitConfig{
				User:       m.Workload.Init.User,
				Cmd:        m.Workload.Init.Cmd,
				Entrypoint: m.Workload.Init.Entrypoint,
			},
			RestartPolicy: RestartPolicyConfig{
				MaxRetries: int(m.Workload.GetRestartPolicy().GetMaxRetries()),
				Policy:     RestartPolicy(m.Workload.GetRestartPolicy().GetPolicy()),
			},
		},
	}
}

func MachineConfigToProto(m MachineConfig) *proto.MachineConfig {
	return &proto.MachineConfig{
		Guest: &proto.GuestConfig{
			Vcpus:    m.Guest.VCpus,
			Cpus:     m.Guest.Cpus,
			MemoryMb: m.Guest.MemoryMB,
		},
		Workload: &proto.Workload{
			Kind:  string(m.Workload.Kind),
			Image: m.Workload.Image,
			Env:   m.Workload.Env,
			Init: &proto.InitConfig{
				User:       m.Workload.Init.User,
				Cmd:        m.Workload.Init.Cmd,
				Entrypoint: m.Workload.Init.Entrypoint,
			},
			RestartPolicy: &proto.RestartPolicy{
				MaxRetries: int32(m.Workload.RestartPolicy.MaxRetries),
				Policy:     string(m.Workload.RestartPolicy.Policy),
			},
		},
	}
}

type MachineTemplate struct {
	VCpus         int64 `json:"cpus"`
	VCpuFrequency int64 `json:"vcpu_frequency"` // MHz
	MemoryMB      int64 `json:"memory_mb"`
}

type MachineTemplates map[string]MachineTemplate

const ECO_CPU_FREQUENCY = 250          // MHz
const PERFORMANCE_CPU_FREQUENCY = 2400 // MHz

var DefaultMachineTemplates = MachineTemplates{
	"eco-1-256": {
		VCpus:         1,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      256,
	},
	"eco-1-512": {
		VCpus:         1,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      512,
	},
	"eco-1-1024": {
		VCpus:         1,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"eco-1-2048": {
		VCpus:         1,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"eco-2-512": {
		VCpus:         2,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      512,
	},
	"eco-2-1024": {
		VCpus:         2,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"eco-2-2048": {
		VCpus:         2,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"eco-2-4096": {
		VCpus:         2,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      4096,
	},
	"eco-4-1024": {
		VCpus:         4,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"eco-4-2048": {
		VCpus:         4,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"eco-4-4096": {
		VCpus:         4,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      4096,
	},
	"eco-4-8192": {
		VCpus:         4,
		VCpuFrequency: ECO_CPU_FREQUENCY,
		MemoryMB:      8192,
	},
	"performance-1-1024": {
		VCpus:         1,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      256,
	},
	"performance-1-2048": {
		VCpus:         1,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      512,
	},
	"performance-1-4096": {
		VCpus:         1,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"performance-1-8192": {
		VCpus:         1,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"performance-2-2048": {
		VCpus:         2,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      512,
	},
	"performance-2-4096": {
		VCpus:         2,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"performance-2-8192": {
		VCpus:         2,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"performance-2-16384": {
		VCpus:         2,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      4096,
	},
	"performance-4-4096": {
		VCpus:         4,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      1024,
	},
	"performance-4-8192": {
		VCpus:         4,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      2048,
	},
	"performance-4-16384": {
		VCpus:         4,
		VCpuFrequency: PERFORMANCE_CPU_FREQUENCY,
		MemoryMB:      4096,
	},
}
