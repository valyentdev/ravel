package instance

import (
	"time"

	"github.com/valyentdev/ravel/api"
)

const (
	InstanceStatusStopped    InstanceStatus = "stopped"
	InstanceStatusStopping   InstanceStatus = "stopping"
	InstanceStatusStarting   InstanceStatus = "starting"
	InstanceStatusRunning    InstanceStatus = "running"
	InstanceStatusDestroying InstanceStatus = "destroying"
)

type InstanceStatus = string

type InstanceMetadata struct {
	MachineId      string `json:"machine_id"`
	MachineVersion string `json:"machine_version"`
}

type Instance struct {
	Id        string           `json:"id"`
	Metadata  InstanceMetadata `json:"metadata"`
	Config    InstanceConfig   `json:"config"`
	ImageRef  string           `json:"image_ref"`
	Network   NetworkingConfig `json:"network"`
	State     State            `json:"state"`
	CreatedAt time.Time        `json:"created_at"`
}

type State struct {
	Status     InstanceStatus `json:"status"`
	ExitResult *ExitResult    `json:"exit_result,omitempty"`
}

type InstanceConfig struct {
	Image      string              `json:"image"`
	Guest      InstanceGuestConfig `json:"guest"`
	Init       api.InitConfig      `json:"init"`
	Env        []string            `json:"env"`
	StopConfig api.StopConfig      `json:"stop_config"`
}

type InstanceGuestConfig struct {
	MemoryMB int `json:"memory_mb"`
	VCpus    int `json:"vcpus"` // number of virtual CPUs (correspond to vm vcpus)
	Cpus     int `json:"cpus"`  // in MHz
}
