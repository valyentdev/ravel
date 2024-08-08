package core

import "time"

type ReservationStatus string

const (
	ReservationStatusDangling  ReservationStatus = "dangling"
	ReservationStatusSuspended ReservationStatus = "suspended"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
)

type Reservation struct {
	Id              string            `json:"id"`
	Cpus            int               `json:"cpus"`   // in MHz
	Memory          int               `json:"memory"` // in MB
	LocalIPV4Subnet string            `json:"local_ipv4_subnet"`
	Status          ReservationStatus `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
}
