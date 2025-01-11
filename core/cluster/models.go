package cluster

import (
	"time"

	"github.com/valyentdev/ravel/api"
)

type Machine struct {
	Id             string     `json:"id"`
	Namespace      string     `json:"namespace"`
	FleetId        string     `json:"fleet_id"`
	InstanceId     string     `json:"instance_id"`
	MachineVersion string     `json:"machine_version"`
	Node           string     `json:"node"`
	Region         string     `json:"region"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DestroyedAt    *time.Time `json:"destroyed_at"`
}

type MachineInstance struct {
	Id                   string             `json:"id"`
	Node                 string             `json:"node"`
	Namespace            string             `json:"namespace"`
	MachineId            string             `json:"machine_id"`
	MachineVersion       string             `json:"machine_version"`
	Status               api.MachineStatus  `json:"status"`
	Events               []api.MachineEvent `json:"events"`
	LocalIPV4            string             `json:"local_ipv4"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
	EnableMachineGateway bool               `json:"enable_machine_gateway"`
}
