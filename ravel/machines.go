package ravel

import (
	"context"
	"crypto/rand"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/oklog/ulid"

	"github.com/alexisbouchez/ravel/api"
	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/core/cluster"
	"github.com/alexisbouchez/ravel/core/config"
	"github.com/alexisbouchez/ravel/core/registry"
	"github.com/alexisbouchez/ravel/internal/id"
)

func getResources(m config.MachineResourcesTemplates, vcpus int, memory int) (api.Resources, error) {
	for _, c := range m.Combinations {
		if c.VCpus == vcpus {
			for _, mc := range c.MemoryConfigs {
				if mc == memory {
					return api.Resources{
						MemoryMB: mc,
						CpusMHz:  m.VCPUFrequency * vcpus,
					}, nil
				}
			}
		}
	}
	return api.Resources{}, errdefs.NewInvalidArgument("Invalid vcpus and memory config")
}

func (r *Ravel) CreateMachine(ctx context.Context, namespace string, fleet string, createOptions api.CreateMachinePayload) (*api.Machine, error) {
	// Validate metadata if provided
	if err := ValidateMetadata(createOptions.Metadata); err != nil {
		return nil, err
	}

	f, err := r.GetFleet(ctx, namespace, fleet)
	if err != nil {
		return nil, err
	}

	config := createOptions.Config

	ref, err := registry.Parse(config.Image)
	if err != nil {
		return nil, errdefs.NewInvalidArgument("Invalid image ref")
	}

	imageRef, err := registry.CheckImageRef(ctx, ref, r.config.Registries)
	if err != nil {
		return nil, errdefs.NewInvalidArgument("Failed to check image ref")
	}

	slog.Debug("Image ref checked", "imageRef", imageRef)

	if r.config.Server.MainRegistry == ref.Domain && r.config.Server.NamespacedRegistry {
		parts := strings.Split(ref.Repository, "/")
		if len(parts) != 2 {
			return nil, errdefs.NewInvalidArgument("Invalid image ref")
		}

		regNS := parts[0]

		if regNS != namespace {
			return nil, errdefs.NewInvalidArgument("Invalid image ref")
		}
	}

	config.Image = imageRef

	// Resolve secrets and inject them as environment variables
	if err := r.resolveSecrets(ctx, namespace, &config); err != nil {
		return nil, err
	}

	// Validate volumes
	if err := validateVolumes(config.Workload.Volumes); err != nil {
		return nil, err
	}

	// Validate private networks
	if err := validatePrivateNetworks(config.Workload.PrivateNetworks); err != nil {
		return nil, err
	}

	ctx = context.Background() // from here we begin to use background context to avoid cancellation of the context passed in and data loss

	versionId := ulid.MustNew(ulid.Now(), rand.Reader).String()
	machine := cluster.Machine{
		Id:             id.Generate(),
		Namespace:      f.Namespace,
		FleetId:        f.Id,
		InstanceId:     id.Generate(),
		MachineVersion: versionId,
		Region:         createOptions.Region,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Metadata:       createOptions.Metadata,
	}

	cputemplate, ok := r.vcpusTemplates[createOptions.Config.Guest.CpuKind]
	if !ok {
		return nil, errdefs.NewInvalidArgument("Invalid CPU kind")
	}

	resources, err := getResources(cputemplate, createOptions.Config.Guest.Cpus, createOptions.Config.Guest.MemoryMB)
	if err != nil {
		return nil, err
	}

	mv := api.MachineVersion{
		Id:        versionId,
		MachineId: machine.Id,
		Namespace: machine.Namespace,
		Config:    config,
		Resources: resources,
	}

	nodeId, err := r.o.PrepareAllocation(ctx, machine.Region, machine.Id, resources)
	if err != nil {
		return nil, err
	}

	machine.Node = nodeId

	err = r.State.CreateMachine(machine, mv)
	if err != nil {
		return nil, err
	}

	err = r.o.PutMachine(ctx, nodeId, &machine, mv, !createOptions.SkipStart, createOptions.EnableMachineGateway)
	if err != nil {
		return nil, err
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
		Status:         api.MachineStatusCreated,
	}, nil
}

func (r *Ravel) StartMachine(ctx context.Context, ns, fleet, machineId string) error {
	machine, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	return r.o.StartMachineInstance(ctx, machine)
}

func (r *Ravel) StopMachine(ctx context.Context, ns, fleet, machineId string, stopConfig *api.StopConfig) error {
	machine, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	return r.o.StopMachineInstance(ctx, machine, stopConfig)
}

func (r *Ravel) MachineExec(ctx context.Context, ns, fleet, machineId string, execOpts *api.ExecOptions) (*api.ExecResult, error) {
	machine, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return nil, err
	}

	return r.o.MachineExec(ctx, machine, execOpts)
}

func (r *Ravel) ListMachines(ctx context.Context, ns, fleet string, includeDestroyed bool) ([]api.Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return nil, err
	}

	return r.State.ListAPIMachines(ctx, ns, f.Id, includeDestroyed)
}

func (r *Ravel) DestroyMachine(ctx context.Context, ns, fleet, machineId string, force bool) error {
	m, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	return r.o.DestroyMachine(ctx, m, force)
}

func (r *Ravel) getMachine(ctx context.Context, ns, fleet, machineId string, showDestroyed bool) (cluster.Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return cluster.Machine{}, err
	}

	return r.State.GetMachine(ctx, ns, f.Id, machineId, showDestroyed)
}

func (r *Ravel) GetMachine(ctx context.Context, ns, fleet, machineId string) (*api.Machine, error) {
	f, err := r.GetFleet(ctx, ns, fleet)
	if err != nil {
		return nil, err
	}

	return r.State.GetAPIMachine(ctx, ns, f.Id, machineId)
}

func (r *Ravel) ListMachineVersions(ctx context.Context, ns, fleet, machineId string) ([]api.MachineVersion, error) {
	_, err := r.getMachine(ctx, ns, fleet, machineId, true)
	if err != nil {
		return nil, err
	}
	return r.State.ListMachineVersions(ctx, machineId)
}

func (r *Ravel) GetMachineLogsRaw(ctx context.Context, ns, fleet, machineId string, follow bool) (io.ReadCloser, error) {
	m, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return nil, err
	}

	return r.o.GetMachineLogsRaw(ctx, m, follow)
}

func (r *Ravel) ListMachineEvents(ctx context.Context, ns, fleet, machineId string) ([]api.MachineEvent, error) {
	m, err := r.getMachine(ctx, ns, fleet, machineId, true)
	if err != nil {
		return nil, err
	}

	return r.State.ListMachineEvents(ctx, m.Id)
}

type waitOpt struct {
	instanceId string
	timeout    time.Duration
}

type WaitOpt func(*waitOpt)

func WithInstanceId(instanceId string) WaitOpt {
	return func(o *waitOpt) {
		o.instanceId = instanceId
	}
}

func WithTimeout(seconds int) WaitOpt {
	return func(o *waitOpt) {
		o.timeout = time.Duration(seconds) * time.Second
	}
}

func validateWaitStatus(status api.MachineStatus) error {
	switch status {
	case api.MachineStatusRunning, api.MachineStatusStopped, api.MachineStatusDestroyed:
		return nil
	default:
		return errdefs.NewInvalidArgument("Invalid status")
	}
}

func (r *Ravel) WaitMachineStatus(ctx context.Context, ns, fleet, machineId string, status api.MachineStatus, timeout uint) error {
	m, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	if err := validateWaitStatus(status); err != nil {
		return err
	}

	return r.o.WaitMachine(ctx, m, status, timeout)
}

func (r *Ravel) EnableMachineGateway(ctx context.Context, ns, fleet, machineId string) error {
	m, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	return r.o.EnableMachineGateway(ctx, m)
}

func (r *Ravel) DisableMachineGateway(ctx context.Context, ns, fleet, machineId string) error {
	m, err := r.getMachine(ctx, ns, fleet, machineId, false)
	if err != nil {
		return err
	}

	return r.o.DisableMachineGateway(ctx, m)
}

// UpdateMachineMetadata updates the metadata for a machine
func (r *Ravel) UpdateMachineMetadata(ctx context.Context, namespace, fleet, machineId string, metadata api.Metadata) (*api.Machine, error) {
	// Validate metadata
	if err := ValidateMetadata(&metadata); err != nil {
		return nil, err
	}

	// Get the machine to ensure it exists
	m, err := r.getMachine(ctx, namespace, fleet, machineId, false)
	if err != nil {
		return nil, err
	}

	// Update the metadata in the state layer
	if err := r.State.UpdateMachineMetadata(ctx, m.Id, metadata); err != nil {
		return nil, err
	}

	// Return the updated machine
	return r.State.GetAPIMachine(ctx, namespace, fleet, machineId)
}

// validateVolumes validates volume mount configurations
func validateVolumes(volumes []api.VolumeMount) error {
	if len(volumes) == 0 {
		return nil
	}

	if len(volumes) > 10 {
		return errdefs.NewInvalidArgument("Too many volumes: maximum 10 allowed")
	}

	seenNames := make(map[string]bool)
	seenPaths := make(map[string]bool)

	for _, vol := range volumes {
		// Validate name
		if vol.Name == "" {
			return errdefs.NewInvalidArgument("Volume name cannot be empty")
		}
		if seenNames[vol.Name] {
			return errdefs.NewInvalidArgument("Duplicate volume name: " + vol.Name)
		}
		seenNames[vol.Name] = true

		// Validate path
		if vol.Path == "" {
			return errdefs.NewInvalidArgument("Volume path cannot be empty")
		}
		if vol.Path[0] != '/' {
			return errdefs.NewInvalidArgument("Volume path must be absolute: " + vol.Path)
		}
		if seenPaths[vol.Path] {
			return errdefs.NewInvalidArgument("Duplicate volume path: " + vol.Path)
		}
		seenPaths[vol.Path] = true
	}

	return nil
}

// validatePrivateNetworks validates private network configurations
func validatePrivateNetworks(networks []api.PrivateNetwork) error {
	if len(networks) == 0 {
		return nil
	}

	if len(networks) > 5 {
		return errdefs.NewInvalidArgument("Too many private networks: maximum 5 allowed")
	}

	seenNames := make(map[string]bool)

	for _, network := range networks {
		// Validate name
		if network.Name == "" {
			return errdefs.NewInvalidArgument("Private network name cannot be empty")
		}
		if seenNames[network.Name] {
			return errdefs.NewInvalidArgument("Duplicate private network name: " + network.Name)
		}
		seenNames[network.Name] = true

		// Validate IP (must be in CIDR notation)
		if network.IP == "" {
			return errdefs.NewInvalidArgument("Private network IP cannot be empty")
		}
		// Simple CIDR validation - check for /prefix
		if !strings.Contains(network.IP, "/") {
			return errdefs.NewInvalidArgument("Private network IP must be in CIDR notation (e.g., 10.0.1.2/24): " + network.IP)
		}
	}

	return nil
}

// resolveSecrets resolves secret references and injects them as environment variables
func (r *Ravel) resolveSecrets(ctx context.Context, namespace string, config *api.MachineConfig) error {
	if len(config.Workload.Secrets) == 0 {
		return nil
	}

	for _, secretRef := range config.Workload.Secrets {
		// Get the secret value from the database
		value, err := r.State.Queries.GetSecretValue(ctx, namespace, secretRef.Name)
		if err != nil {
			return errdefs.NewInvalidArgument("Secret not found: " + secretRef.Name)
		}

		// Inject the secret as an environment variable
		envVar := secretRef.EnvVar + "=" + value
		config.Workload.Env = append(config.Workload.Env, envVar)
	}

	return nil
}
