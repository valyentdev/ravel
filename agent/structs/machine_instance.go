package structs

import (
	"time"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/cluster"
	"github.com/valyentdev/ravel/core/instance"
)

type MachineInstanceState struct {
	DesiredStatus api.MachineStatus  `json:"desired_status"`
	Status        api.MachineStatus  `json:"status"`
	Restarts      int                `json:"restarts"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	LocalIPV4     string             `json:"local_ipv4"`
	LastEvents    []api.MachineEvent `json:"last_events"`
}

type MachineInstance struct {
	Machine cluster.Machine      `json:"machine"`
	Version api.MachineVersion   `json:"version"`
	State   MachineInstanceState `json:"state"`
}

func (mi *MachineInstance) InstanceOptions() instance.InstanceOptions {
	return instance.InstanceOptions{
		Id: mi.Machine.InstanceId,
		Metadata: instance.InstanceMetadata{
			MachineId:      mi.Machine.InstanceId,
			MachineVersion: mi.Version.Id,
		},
		Config: instance.InstanceConfig{
			Image: mi.Version.Config.Image,
			Guest: instance.InstanceGuestConfig{
				MemoryMB: mi.Version.Resources.MemoryMB,
				CpusMHz:  mi.Version.Resources.CpusMHz,
				VCpus:    mi.Version.Config.Guest.Cpus,
			},
			Init: mi.Version.Config.Workload.Init,
			Stop: mi.Version.Config.StopConfig,
			Env:  mi.Version.Config.Workload.Env,
		},
	}
}

func (mi *MachineInstance) ClusterInstance() cluster.MachineInstance {
	return cluster.MachineInstance{
		Id:             mi.Machine.InstanceId,
		Node:           mi.Machine.Node,
		Namespace:      mi.Machine.Namespace,
		MachineId:      mi.Machine.Id,
		MachineVersion: mi.Version.Id,
		Events:         mi.State.LastEvents,
		Status:         mi.State.Status,
		LocalIPV4:      mi.State.LocalIPV4,
		CreatedAt:      mi.State.CreatedAt,
		UpdatedAt:      mi.State.UpdatedAt,
	}
}
