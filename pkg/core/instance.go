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
	Id             string        `json:"id"`
	Namespace      string        `json:"namespace"`
	FleetId        string        `json:"fleet_id"`
	NodeId         string        `json:"node_id"`
	MachineId      string        `json:"machine_id"`
	ReservationId  string        `json:"reservation_id"`
	DesiredStatus  MachineStatus `json:"desired_status"`
	Restarts       int           `json:"restarts"`
	MachineVersion string        `json:"machine_version"`
	Config         MachineConfig `json:"config"`
	CreatedAt      time.Time     `json:"created_at"`
	Prepared       bool          `json:"prepared"`
	Destroyed      bool          `json:"destroyed"`
	DestroyedAt    time.Time     `json:"destroyed_at"`
}
