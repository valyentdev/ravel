package agent

import (
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MachineConfigToProto(config core.MachineConfig) *proto.MachineConfig {
	return &proto.MachineConfig{
		Guest: &proto.GuestConfig{
			Cpus:     config.Guest.Cpus,
			MemoryMb: config.Guest.MemoryMB,
		},
		Workload: &proto.Workload{
			Kind:  string(config.Workload.Kind),
			Image: config.Workload.Image,
			RestartPolicy: &proto.RestartPolicy{
				MaxRetries: int32(config.Workload.RestartPolicy.MaxRetries),
				Policy:     string(config.Workload.RestartPolicy.Policy),
			},
		},
	}
}

func MachineConfigFromProto(config *proto.MachineConfig) core.MachineConfig {

	restartPolicy := config.Workload.GetRestartPolicy()
	if restartPolicy != nil {
		restartPolicy = &proto.RestartPolicy{}
	}

	return core.MachineConfig{
		Guest: core.GuestConfig{
			VCpus:    config.Guest.Cpus,
			Cpus:     config.Guest.Cpus,
			MemoryMB: config.Guest.MemoryMb,
		},
		Workload: core.Workload{
			Kind:          core.WorkloadKind(config.Workload.Kind),
			Image:         config.Workload.Image,
			RestartPolicy: core.RestartPolicyConfig{},
		},
	}
}

func InstanceToProto(m *core.Instance) *proto.Instance {
	return &proto.Instance{
		Id:        m.Id,
		MachineId: m.MachineId,
		Config:    MachineConfigToProto(m.Config),
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}
