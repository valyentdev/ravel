package placement

import "github.com/valyentdev/ravel/pkg/core"

type MachinePlacementRequest struct {
	Region        string `json:"region"`
	ReservationId string `json:"reservation_id"`
	Cpus          int    `json:"cpus"`   // in MHz
	Memory        int    `json:"memory"` // in MB
}

type MachinePlacementResponse struct {
	NodeId          string
	CpuFrequency    int // in MHz
	CpuCount        int // Correspond to the number of cores * number of threads per core
	Allocatable     core.Resources
	AllocatedBefore core.Resources
	AllocatedAfter  core.Resources
}

func (r MachinePlacementResponse) GetScore() float64 {
	cpuUtilization := float64(r.AllocatedAfter.Cpus) / float64(r.Allocatable.Cpus)
	memoryUtilization := float64(r.AllocatedAfter.Memory) / float64(r.Allocatable.Memory)

	idealRatio := float64(r.Allocatable.Cpus) / float64(r.Allocatable.Memory)
	currentRatio := float64(r.AllocatedAfter.Cpus) / float64(r.AllocatedAfter.Memory)

	ratioScore := 1 - (idealRatio-currentRatio)/idealRatio

	return 0.4*cpuUtilization + 0.4*memoryUtilization + 0.2*ratioScore

}
