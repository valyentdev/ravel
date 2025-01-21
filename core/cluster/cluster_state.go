package cluster

import (
	"context"

	"github.com/valyentdev/ravel/api"
)

type Queries interface {

	/* Used by the API server */
	CreateGateway(ctx context.Context, gateway api.Gateway) error
	DeleteGateway(ctx context.Context, id string) error
	DeleteFleetGateways(ctx context.Context, fleetId string) error

	CreateMachineVersion(ctx context.Context, mv api.MachineVersion) error
	CreateMachine(ctx context.Context, m Machine, mv api.MachineVersion) error
	UpdateMachine(ctx context.Context, m Machine) error
	DestroyMachine(ctx context.Context, id string) error

	ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error)
	GetAPIMachine(ctx context.Context, namespace string, fleetId string, machineId string) (*api.Machine, error)
	DestroyNamespaceData(ctx context.Context, namespace string) error

	GetNode(ctx context.Context, id string) (api.Node, error)
	ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error)
	ListNodes(ctx context.Context) ([]api.Node, error)
	ListRegions(ctx context.Context) ([]string, error)

	/* Used by raveld */
	UpsertNode(ctx context.Context, node api.Node) error
	UpsertInstance(ctx context.Context, i MachineInstance) error
}

type ClusterState interface {
	Queries
	BeginTx(ctx context.Context) (TX, error)
}

type TX interface {
	Queries
	Commit(context.Context) error
	Rollback(context.Context) error
}
