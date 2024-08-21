package api

import (
	"time"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/pkg/core"
)

type Machine struct {
	Id             string             `json:"id"`
	Namespace      string             `json:"namespace"`
	FleetId        string             `json:"fleet"`
	InstanceId     string             `json:"instance_id"`
	MachineVersion ulid.ULID          `json:"machine_version"`
	Region         string             `json:"region"`
	Config         core.MachineConfig `json:"config"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	State          core.MachineStatus `json:"state"`
}

type Namespace struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
