package cluster

import (
	"context"
	"errors"

	"github.com/alexisbouchez/ravel/api"
)

// NewNoop returns a ClusterState that performs no cross-node state
// replication. Queries return empty results or ErrNoClusterState; mutations
// are discarded. Intended as a single-node fallback.
func NewNoop() ClusterState {
	return noopState{}
}

var ErrNoClusterState = errors.New("cluster state is not configured")

type noopState struct{}

func (noopState) CreateGateway(ctx context.Context, gateway api.Gateway) error { return nil }
func (noopState) DeleteGateway(ctx context.Context, id string) error           { return nil }
func (noopState) DeleteFleetGateways(ctx context.Context, fleetId string) error {
	return nil
}

func (noopState) CreateMachineVersion(ctx context.Context, mv api.MachineVersion) error {
	return nil
}
func (noopState) CreateMachine(ctx context.Context, m Machine, mv api.MachineVersion) error {
	return nil
}
func (noopState) UpdateMachine(ctx context.Context, m Machine) error { return nil }
func (noopState) UpdateMachineMetadata(ctx context.Context, machineID string, metadata api.Metadata) error {
	return nil
}
func (noopState) UpdateFleetMetadata(ctx context.Context, fleetID string, metadata api.Metadata) error {
	return nil
}
func (noopState) DestroyMachine(ctx context.Context, id string) error { return nil }

func (noopState) ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error) {
	return nil, nil
}
func (noopState) GetAPIMachine(ctx context.Context, namespace string, fleetId string, machineId string) (*api.Machine, error) {
	return nil, ErrNoClusterState
}
func (noopState) DestroyNamespaceData(ctx context.Context, namespace string) error { return nil }

func (noopState) GetNode(ctx context.Context, id string) (api.Node, error) {
	return api.Node{}, ErrNoClusterState
}
func (noopState) ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error) {
	return nil, nil
}
func (noopState) ListNodes(ctx context.Context) ([]api.Node, error) { return nil, nil }
func (noopState) ListRegions(ctx context.Context) ([]string, error) { return nil, nil }

func (noopState) UpsertNode(ctx context.Context, node api.Node) error         { return nil }
func (noopState) UpsertInstance(ctx context.Context, i MachineInstance) error { return nil }

func (noopState) BeginTx(ctx context.Context) (TX, error) {
	return noopTx{}, nil
}

type noopTx struct{ noopState }

func (noopTx) Commit(context.Context) error   { return nil }
func (noopTx) Rollback(context.Context) error { return nil }
