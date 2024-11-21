package structs

import (
	"time"

	"github.com/valyentdev/ravel/api"
)

type AllocationStatus string

const (
	AllocationStatusDangling  AllocationStatus = "dangling"
	AllocationStatusSuspended AllocationStatus = "suspended"
	AllocationStatusConfirmed AllocationStatus = "confirmed"
)

type Allocation struct {
	Id        string           `json:"id"`
	Resources api.Resources    `json:"resources"`
	Status    AllocationStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}
