package core

import (
	"time"
)

type Origin string

const (
	OriginUser  Origin = "user"
	OriginRavel Origin = "ravel"
)

type InstanceStatus = MachineStatus

type Instance struct {
	Id             string         `json:"id"`
	Namespace      string         `json:"namespace"`
	FleetId        string         `json:"fleet_id"`
	NodeId         string         `json:"node_id"`
	MachineId      string         `json:"machine_id"`
	MachineVersion string         `json:"machine_version"`
	ReservationId  string         `json:"reservation_id"`
	DesiredStatus  InstanceStatus `json:"desired_status"`
	Restarts       int            `json:"restarts"`
	Config         InstanceConfig `json:"config"`
	CreatedAt      time.Time      `json:"created_at"`
	Status         InstanceStatus `json:"status"`
}

type InstanceConfig struct {
	Guest      InstanceGuestConfig
	Workload   Workload
	StopConfig StopConfig
}

type InstanceGuestConfig struct {
	MemoryMB int `json:"memory_mb"`
	VCpus    int `json:"vcpus"` // number of virtual CPUs (correspond to ch vcpus process threads)
	Cpus     int `json:"cpus"`  // in MHz
}
