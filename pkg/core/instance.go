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
	Id        string `json:"id"`
	Namespace string `json:"namespace"`
	NodeId    string `json:"node_id"`
	MachineId string `json:"machine_id"`
	FleetId   string `json:"fleet_id"`

	Config         InstanceConfig `json:"config"`
	MachineVersion string         `json:"machine_version"`
	LocalIPV4      string         `json:"local_ipv4"`
	ReservationId  string         `json:"reservation_id"`

	State     InstanceState `json:"state"`
	CreatedAt time.Time     `json:"created_at"`
}

type InstanceState struct {
	DesiredStatus InstanceStatus `json:"desired_status"`
	Status        InstanceStatus `json:"status"`
	Restarts      int            `json:"restarts"`
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
