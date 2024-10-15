package structs

import (
	"time"

	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
)

type ReservationStatus string

const (
	ReservationStatusDangling  ReservationStatus = "dangling"
	ReservationStatusSuspended ReservationStatus = "suspended"
	ReservationStatusConfirmed ReservationStatus = "confirmed"
)

type Reservation struct {
	Id              string                     `json:"id"`
	Resources       core.Resources             `json:"resources"`
	LocalIPV4Subnet networking.LocalIPV4Subnet `json:"local_ipv4_subnet"`
	Status          ReservationStatus          `json:"status"`
	CreatedAt       time.Time                  `json:"created_at"`
}
