package api

import (
	"time"
)

type Namespace struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type FleetStatus string

const (
	FleetStatusActive    FleetStatus = "active"
	FleetStatusDestroyed FleetStatus = "destroyed"
)

type Fleet struct {
	Id        string      `json:"id"`
	Namespace string      `json:"namespace"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
	Status    FleetStatus `json:"status"`
}

type MachineVersion struct {
	Id        string        `json:"id"`
	MachineId string        `json:"machine_id"`
	Namespace string        `json:"namespace"`
	Config    MachineConfig `json:"config"`
	Resources Resources     `json:"resources"`
}

type LogEntry struct {
	Timestamp  int64  `json:"timestamp,omitempty"`
	InstanceId string `json:"instance_id,omitempty"`
	Source     string `json:"source,omitempty"`
	Level      string `json:"level,omitempty"`
	Message    string `json:"message,omitempty"`
}

type Resources struct {
	CpusMHz  int `json:"cpus_mhz" toml:"cpus_mhz"`   // in MHz
	MemoryMB int `json:"memory_mb" toml:"memory_mb"` // in MB
}

func (r *Resources) Sub(other Resources) Resources {
	new := Resources{
		CpusMHz:  r.CpusMHz - other.CpusMHz,
		MemoryMB: r.MemoryMB - other.MemoryMB,
	}
	return new
}

// Add returns a new Resources object which is the sum of the resources.
func (r *Resources) Add(other Resources) Resources {
	new := Resources{
		CpusMHz:  r.CpusMHz + other.CpusMHz,
		MemoryMB: r.MemoryMB + other.MemoryMB,
	}
	return new
}

// GT returns true if the resources are greater than the other resources.
func (r *Resources) GT(other Resources) bool {
	return r.CpusMHz > other.CpusMHz || r.MemoryMB > other.MemoryMB
}
