package cluster

import (
	"context"

	"github.com/valyentdev/ravel/api"
	"github.com/valyentdev/ravel/core/instance"
)

type ClusterState interface {
	CreateGateway(ctx context.Context, gateway api.Gateway) error
	DeleteGateway(ctx context.Context, id string) error
	DeleteFleetGateways(ctx context.Context, fleetId string) error

	CreateMachine(ctx context.Context, m Machine, mv api.MachineVersion) error
	UpdateMachine(ctx context.Context, m Machine) error
	DestroyMachine(ctx context.Context, id string) error

	UpsertInstance(ctx context.Context, i MachineInstance) error
	WatchInstanceStatus(ctx context.Context, machineId string, instanceId string) (<-chan instance.InstanceStatus, error)

	GetNode(ctx context.Context, id string) (api.Node, error)
	UpsertNode(ctx context.Context, node api.Node) error
	ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error)
	ListNodes(ctx context.Context) ([]api.Node, error)

	ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error)
	GetAPIMachine(ctx context.Context, namespace string, fleetId string, machineId string) (*api.Machine, error)

	DestroyNamespaceData(ctx context.Context, namespace string) error
}
