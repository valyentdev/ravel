package ravel

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"github.com/oklog/ulid"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/valyentdev/ravel/pkg/core/api"
	"github.com/valyentdev/ravel/pkg/core/config"
	"github.com/valyentdev/ravel/pkg/ravel/orchestrator"
)

func getInstanceGuestConfig(m config.MachineResourcesTemplates, vcpus int, memory int) (core.InstanceGuestConfig, error) {
	for _, c := range m.Combinations {
		if c.VCpus == vcpus {
			for _, mc := range c.MemoryConfigs {
				if mc == memory {
					return core.InstanceGuestConfig{
						VCpus:    vcpus,
						MemoryMB: memory,
						Cpus:     m.FrequencyByCpu * vcpus,
					}, nil
				}
			}
		}
	}
	return core.InstanceGuestConfig{}, core.NewInvalidArgument("Invalid vcpus and memory config")
}

type CreateMachineOptions struct {
	Region    string             `json:"region"`
	Config    core.MachineConfig `json:"config"`
	SkipStart bool               `json:"skip_start"`
}

func (r *Ravel) CreateMachine(ctx context.Context, namespace string, fleet string, createOptions CreateMachineOptions) (*Machine, error) {
	versionId := ulid.MustNew(ulid.Now(), rand.Reader)
	f, err := r.GetFleet(ctx, namespace, fleet)
	if err != nil {
		return nil, err
	}

	machine := core.Machine{
		Id:             id.Generate(),
		Namespace:      f.Namespace,
		FleetId:        f.Id,
		MachineVersion: versionId,
		Region:         createOptions.Region,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mv := core.MachineVersion{
		Id:        versionId,
		MachineId: machine.Id,
		Config:    createOptions.Config,
	}

	cputemplate, ok := r.vcpusTemplates[createOptions.Config.Guest.CpuKind]
	if !ok {
		return nil, core.NewInvalidArgument("Invalid CPU kind")
	}

	slog.Info("Creating machine", "machine_id", machine.Id, "fleet_id", f.Id, "namespace", namespace)

	guestConfig, err := getInstanceGuestConfig(cputemplate, createOptions.Config.Guest.Cpus, createOptions.Config.Guest.MemoryMB)
	if err != nil {
		return nil, err
	}

	instanceConfig := core.InstanceConfig{
		Guest:      guestConfig,
		Workload:   createOptions.Config.Workload,
		StopConfig: createOptions.Config.StopConfig,
	}

	instance, err := r.o.CreateInstanceForMachine(ctx, namespace, f.Id, orchestrator.CreateInstanceOptions{
		Machine:   machine,
		Config:    instanceConfig,
		SkipStart: createOptions.SkipStart,
	})
	if err != nil {
		return nil, err
	}

	machine.InstanceId = instance.Id
	machine.Node = instance.NodeId

	err = r.state.CreateMachine(machine, mv)
	if err != nil {
		return nil, err // TODO: destroy instance on error here
	}

	return &api.Machine{
		Id:             machine.Id,
		Namespace:      machine.Namespace,
		FleetId:        machine.FleetId,
		InstanceId:     machine.InstanceId,
		MachineVersion: machine.MachineVersion,
		Region:         machine.Region,
		Config:         createOptions.Config,
		CreatedAt:      machine.CreatedAt,
		UpdatedAt:      machine.UpdatedAt,
		State:          core.MachineStatusCreated,
	}, nil
}

func (r *Ravel) StartMachine(ctx context.Context, ns, fleet, machineId string) error {
	machine, err := r.getMachine(ctx, ns, fleet, machineId)
	if err != nil {
		return err
	}

	if machine.Destroyed {
		return core.NewFailedPrecondition("machine is destroyed")
	}

	return r.o.StartMachineInstance(ctx, machine)
}

func (r *Ravel) StopMachine(ctx context.Context, ns, fleet, machineId string, stopConfig *core.StopConfig) error {
	machine, err := r.getMachine(ctx, ns, fleet, machineId)
	if err != nil {
		return err
	}

	if machine.Destroyed {
		return core.NewFailedPrecondition("machine is destroyed")
	}

	return r.o.StopMachineInstance(ctx, machine, stopConfig)
}

func (r *Ravel) ListMachines(ctx context.Context, ns, fleet string, includeDestroyed bool) ([]Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return nil, err
	}

	return r.clusterState.ListAPIMachines(ctx, ns, f.Id, includeDestroyed)
}

func (r *Ravel) DestroyMachine(ctx context.Context, ns, fleet, machineId string, force bool) error {
	m, err := r.getMachine(ctx, ns, fleet, machineId)
	if err != nil {
		return err
	}

	if m.Destroyed {
		return nil
	}

	return r.o.DestroyMachine(ctx, m, force)
}

func (r *Ravel) getMachine(ctx context.Context, ns, fleet, machineId string) (core.Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return core.Machine{}, err
	}

	return r.db.GetMachine(ctx, ns, f.Id, machineId)
}

func (r *Ravel) GetMachine(ctx context.Context, ns, fleet, machineId string) (*Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return nil, err
	}

	slog.Info("Getting machine", "machine_id", machineId, "fleet_id", f.Id, "namespace", ns)

	return r.clusterState.GetAPIMachine(ctx, ns, f.Id, machineId)
}

func (r *Ravel) ListMachineVersions(ctx context.Context, ns, fleet, machineId string) ([]core.MachineVersion, error) {
	_, err := r.getMachine(ctx, ns, fleet, machineId)
	if err != nil {
		return nil, err
	}
	return r.db.ListMachineVersions(ctx, machineId)
}
