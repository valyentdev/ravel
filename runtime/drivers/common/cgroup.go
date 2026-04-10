package common

import (
	"github.com/alexisbouchez/ravel/core/instance"
	"github.com/containerd/cgroups/v3/cgroup2"
)

// GetCgroupMemory returns the cgroup memory configuration for an instance.
func GetCgroupMemory(c *instance.InstanceGuestConfig) *cgroup2.Memory {
	high := int64(c.MemoryMB * 1_000_000)
	max := int64(float64(c.MemoryMB) * 1_000_000 * 1.1) // trigger OOM at 110% of the limit
	return &cgroup2.Memory{
		High: &high,
		Max:  &max,
	}
}

// GetCgroupCPU returns the cgroup CPU configuration for an instance.
func GetCgroupCPU(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.CPU {
	quota := int64(float64(c.CpusMHz) / float64(cpuMhz) * 1_000_000)
	period := uint64(1_000_000)
	max := cgroup2.NewCPUMax(&quota, &period)
	return &cgroup2.CPU{
		Max: max,
	}
}

// GetInstanceResources returns the combined cgroup resources for an instance.
func GetInstanceResources(cpuMhz int64, c *instance.InstanceGuestConfig) *cgroup2.Resources {
	return &cgroup2.Resources{
		CPU:    GetCgroupCPU(cpuMhz, c),
		Memory: GetCgroupMemory(c),
	}
}
