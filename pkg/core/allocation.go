package core

type AllocationStatus string

type Allocation struct {
	Id        string
	Status    string
	Resources Resources
}

const (
	AllocationStatusDangling  AllocationStatus = "dangling"
	AllocationStatusSuspended AllocationStatus = "suspended"
	AllocationStatusConfirmed AllocationStatus = "confirmed"
)
