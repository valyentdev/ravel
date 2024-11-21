package cluster

import (
	"context"

	"github.com/valyentdev/ravel/api"
)

type ClusterState interface {
	CreateGateway(ctx context.Context, gateway api.Gateway) error
	DeleteGateway(ctx context.Context, id string) error

	CreateMachine(ctx context.Context, m Machine, mv api.MachineVersion) error
	UpdateMachine(ctx context.Context, m Machine) error
	GetMachine(ctx context.Context, namespace string, fleetId string, id string, destroyed bool) (*Machine, error)
	ListMachines(ctx context.Context, namespace string, fleet string, destroyed bool) ([]Machine, error)
	DestroyMachine(ctx context.Context, id string) error

	GetInstance(ctx context.Context, id string) (*MachineInstance, error)
	UpsertInstance(ctx context.Context, i MachineInstance) error
	WatchInstance(ctx context.Context, machineId string, instanceId string) (context.CancelFunc, <-chan MachineInstance, error)

	GetNode(ctx context.Context, id string) (api.Node, error)
	UpsertNode(ctx context.Context, node api.Node) error
	ListNodesInRegion(ctx context.Context, region string) ([]api.Node, error)
	ListNodes(ctx context.Context) ([]api.Node, error)

	ListAPIMachines(ctx context.Context, namespace string, fleetId string, includeDestroyed bool) ([]api.Machine, error)
	GetAPIMachine(ctx context.Context, namespace string, fleetId string, machineId string) (*api.Machine, error)
}
