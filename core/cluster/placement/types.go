package placement

import "github.com/valyentdev/ravel/api"

type PlacementRequest struct {
	AllocationId string        `json:"allocation_id"`
	Region       string        `json:"region"`
	Resources    api.Resources `json:"resources"`
}

type PlacementResponse struct {
	NodeId          string        `json:"node_id"`
	Allocatable     api.Resources `json:"allocatable"`
	AllocatedBefore api.Resources `json:"allocated_before"`
	AllocatedAfter  api.Resources `json:"allocated_after"`
}

func (r PlacementResponse) GetScore() float64 {
	cpuUtilization := float64(r.AllocatedAfter.CpusMHz) / float64(r.Allocatable.CpusMHz)
	memoryUtilization := float64(r.AllocatedAfter.MemoryMB) / float64(r.Allocatable.MemoryMB)

	idealRatio := float64(r.Allocatable.MemoryMB) / float64(r.Allocatable.CpusMHz)
	currentRatio := float64(r.AllocatedAfter.MemoryMB) / float64(r.AllocatedAfter.CpusMHz)

	ratioScore := 1 - (idealRatio-currentRatio)/idealRatio

	return 0.4*cpuUtilization + 0.4*memoryUtilization + 0.2*ratioScore

}
