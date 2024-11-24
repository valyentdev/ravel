package api

import (
	"time"

	"github.com/oklog/ulid"
)

type Namespace struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Fleet struct {
	Id        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Destroyed bool      `json:"destroyed"`
}

type MachineVersion struct {
	Id        ulid.ULID     `json:"id"`
	MachineId string        `json:"machine_id"`
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
