package api

import "time"

type Fleet struct {
	Id        string      `json:"id"`
	Namespace string      `json:"namespace"`
	Name      string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
	Status    FleetStatus `json:"status"`
}

type FleetStatus string

const (
	FleetStatusActive    FleetStatus = "active"
	FleetStatusDestroyed FleetStatus = "destroyed"
)

type CreateFleetPayload struct {
	Name string `json:"name"`
}
