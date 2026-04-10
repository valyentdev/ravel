package api

import (
	"time"
)

type Namespace struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
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
