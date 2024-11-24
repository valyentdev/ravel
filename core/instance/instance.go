package instance

import (
	"time"

	"github.com/valyentdev/ravel/api"
)

const (
	InstanceStatusCreated    InstanceStatus = "created"
	InstanceStatusStopped    InstanceStatus = "stopped"
	InstanceStatusStarting   InstanceStatus = "starting"
	InstanceStatusRunning    InstanceStatus = "running"
	InstanceStatusDestroying InstanceStatus = "destroying"
	InstanceStatusDestroyed  InstanceStatus = "destroyed"
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
	Stopping   bool           `json:"stopping"`
	ExitResult *ExitResult    `json:"exit_result,omitempty"`
}

type InstanceConfig struct {
	Image string              `json:"image"`
	Guest InstanceGuestConfig `json:"guest"`
	Init  api.InitConfig      `json:"init"`
	Stop  *api.StopConfig     `json:"stop,omitempty"`
	Env   []string            `json:"env"`
}

type InstanceGuestConfig struct {
	MemoryMB int `json:"memory_mb" minimum:"1"` // in MB
	VCpus    int `json:"vcpus" minimum:"1"`     // number of virtual CPUs (correspond to vm vcpus)
	CpusMHz  int `json:"cpus_mhz" minimum:"1"`  // in MHz
}
