package core

import (
	"time"

	"github.com/oklog/ulid"
)

type Machine struct {
	Id             string    `json:"id"`
	Namespace      string    `json:"namespace"`
	FleetId        string    `json:"fleet_id"`
	InstanceId     string    `json:"instance_id"`
	MachineVersion ulid.ULID `json:"machine_version"`
	Node           string    `json:"node"`
	Region         string    `json:"region"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Destroyed      bool      `json:"destroyed"`
}

type MachineConfig struct {
	Guest       GuestConfig `json:"guest"`
	Workload    Workload    `json:"workload"`
	StopConfig  StopConfig  `json:"stop_config,omitempty"`
	AutoDestroy bool        `json:"auto_destroy,omitempty"`
}

type MachineVersion struct {
	Id        ulid.ULID     `json:"id"`
	MachineId string        `json:"machine_id"`
	Config    MachineConfig `json:"config"`
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
