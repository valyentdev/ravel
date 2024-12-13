package endpoints

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/ravel"
)

type Endpoints struct {
	ravel *ravel.Ravel
}

func New(r *ravel.Ravel) *Endpoints {
	return &Endpoints{ravel: r}
}

func (e *Endpoints) Register(api huma.API) {
	e.registerFleetsEndpoints(api)
	e.registerMachinesEndpoints(api)
	e.registerGatewaysEndpoints(api)
}

func (e *Endpoints) registerFleetsEndpoints(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createFleet",
		Summary:     "Create a fleet",
		Path:        "/ns/{namespace}/fleets",
		Method:      http.MethodPost,
		Tags:        []string{"fleets"},
	}, e.createFleet)

	huma.Register(api, huma.Operation{
		OperationID: "listFleets",
		Summary:     "List fleets",
		Path:        "/ns/{namespace}/fleets",
		Method:      http.MethodGet,
		Tags:        []string{"fleets"},
	}, e.listFleets)

	huma.Register(api, huma.Operation{
		OperationID: "getFleet",
		Summary:     "Get a fleet",
		Path:        "/ns/{namespace}/fleets/{fleet}",
		Method:      http.MethodGet,
		Tags:        []string{"fleets"},
	}, e.getFleet)

	huma.Register(api, huma.Operation{
		OperationID: "destroyFleet",
		Summary:     "Destroy a fleet",
		Path:        "/ns/{namespace}/fleets/{fleet}",
		Method:      http.MethodDelete,
		Tags:        []string{"fleets"},
	}, e.destroyFleet)
}

func (e *Endpoints) registerMachinesEndpoints(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createMachine",
		Summary:     "Create a machine",
		Method:      http.MethodPost,
		Path:        "/ns/{namespace}/fleets/{fleet}/machines",
		Tags:        []string{"machines"},
	}, e.createMachine)

	huma.Register(api, huma.Operation{
		OperationID: "destroyMachine",
		Summary:     "Destroy a machine",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}",
		Method:      http.MethodDelete,
		Tags:        []string{"machines"},
	}, e.destroyMachine)

	huma.Register(api, huma.Operation{
		OperationID: "listMachines",
		Summary:     "List machines",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.listMachines)

	huma.Register(api, huma.Operation{
		OperationID: "getMachine",
		Summary:     "Get a machine",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.getMachine)

	huma.Register(api, huma.Operation{
		OperationID: "startMachine",
		Summary:     "Start a machine",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/start",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.startMachine)

	huma.Register(api, huma.Operation{
		OperationID: "stopMachine",
		Summary:     "Stop a machine",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/stop",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.stopMachine)

	huma.Register(api, huma.Operation{
		OperationID: "machineExec",
		Summary:     "Execute a command inside a machine",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/exec",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.machineExec)

	huma.Register(api, huma.Operation{
		OperationID: "listMachineVersions",
		Summary:     "List machine versions",
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/versions",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.listMachineVersions)

	huma.Register(api, huma.Operation{
		OperationID: "listMachineEvents",
		Summary:     "List machine events",
		Method:      http.MethodGet,
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/events",
		Tags:        []string{"machines"},
	}, e.listMachineEvents)

	huma.Register(api, huma.Operation{
		OperationID: "getMachineLogs",
		Summary:     "Get machine logs",
		Method:      http.MethodGet,
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/logs",
		Tags:        []string{"machines"},
	}, e.getMachineLogs)

	huma.Register(api, huma.Operation{
		OperationID: "waitMachineStatus",
		Summary:     "Wait for a machine to reach a given status",
		Method:      http.MethodGet,
		Path:        "/ns/{namespace}/fleets/{fleet}/machines/{machine_id}/wait",
		Tags:        []string{"machines"},
	}, e.waitMachineStatus)
}

func (e *Endpoints) registerGatewaysEndpoints(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createGateway",
		Summary:     "Create a gateway",
		Method:      http.MethodPost,
		Path:        "/ns/{namespace}/gateways",
		Tags:        []string{"gateways"},
	}, e.createGateway)

	huma.Register(api, huma.Operation{
		OperationID: "listGateways",
		Summary:     "List gateways",
		Method:      http.MethodGet,
		Path:        "/ns/{namespace}/gateways",
		Tags:        []string{"gateways"},
	}, e.listGateways)

	huma.Register(api, huma.Operation{
		OperationID: "getGateway",
		Summary:     "Get a gateway",
		Method:      http.MethodGet,
		Path:        "/ns/{namespace}/gateways/{gateway}",
		Tags:        []string{"gateways"},
	}, e.getGateway)

	huma.Register(api, huma.Operation{
		OperationID: "destroyGateway",
		Summary:     "Destroy a gateway",
		Method:      http.MethodDelete,
		Path:        "/ns/{namespace}/gateways/{gateway}",
		Tags:        []string{"gateways"},
	}, e.destroyGateway)

}
