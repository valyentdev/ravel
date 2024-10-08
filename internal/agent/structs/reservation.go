package structs

import (
	"time"

	"github.com/valyentdev/ravel/internal/networking"
)

type ReservationStatus string

const (
	ReservationStatusDangling  ReservationStatus = "dangling"
	ReservationStatusSuspended ReservationStatus = "suspended"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
)

type Reservation struct {
	Id              string                     `json:"id"`
	Cpus            int                        `json:"cpus"`   // in MHz
	Memory          int                        `json:"memory"` // in MB
	LocalIPV4Subnet networking.LocalIPV4Subnet `json:"local_ipv4_subnet"`
	Status          ReservationStatus          `json:"status"`
	CreatedAt       time.Time                  `json:"created_at"`
}
