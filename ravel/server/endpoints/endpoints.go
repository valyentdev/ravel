package endpoints

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexisbouchez/ravel/api/errdefs"
	"github.com/alexisbouchez/ravel/ravel"
	"github.com/danielgtaylor/huma/v2"
)

type Endpoints struct {
	ravel *ravel.Ravel
}

func (e *Endpoints) log(msg string, err error) {
	var rerr *errdefs.RavelError
	if errors.As(err, &rerr) {
		if errdefs.IsUnknown(err) || errdefs.IsInternal(err) {
			slog.Error(msg, "error", err)
		}
	} else {
		slog.Error(msg, "error", err)
	}
}

func New(r *ravel.Ravel) *Endpoints {
	return &Endpoints{ravel: r}
}

func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listNodes",
		Path:        "/nodes",
		Method:      http.MethodGet,
		Tags:        []string{"nodes"},
		Summary:     "List nodes",
	}, e.listNodes)

	huma.Register(api, huma.Operation{
		OperationID: "createNamespace",
		Path:        "/namespaces",
		Method:      "POST",
		Summary:     "Create a new namespace.",
		Tags:        []string{"namespaces"},
	}, e.createNamespace)

	huma.Register(api, huma.Operation{
		OperationID: "listNamespaces",
		Path:        "/namespaces",
		Method:      http.MethodGet,
		Tags:        []string{"namespaces"},
		Summary:     "List namespaces",
	}, e.listNamespaces)

	huma.Register(api, huma.Operation{
		OperationID: "getNamespace",
		Path:        "/namespaces/{namespace}",
		Method:      http.MethodGet,
		Tags:        []string{"namespaces"},
		Summary:     "Get a namespace",
	}, e.getNamespace)

	huma.Register(api, huma.Operation{
		OperationID: "destroyNamespace",
		Path:        "/namespaces/{namespace}",
		Method:      http.MethodDelete,
		Tags:        []string{"namespaces"},
		Summary:     "Delete a namespace",
	}, e.destroyNamespace)

	huma.Register(api, huma.Operation{
		OperationID: "createSecret",
		Summary:     "Create a secret",
		Method:      http.MethodPost,
		Path:        "/namespaces/{namespace}/secrets",
		Tags:        []string{"secrets"},
	}, e.createSecret)

	huma.Register(api, huma.Operation{
		OperationID: "listSecrets",
		Summary:     "List secrets",
		Method:      http.MethodGet,
		Path:        "/namespaces/{namespace}/secrets",
		Tags:        []string{"secrets"},
	}, e.listSecrets)

	huma.Register(api, huma.Operation{
		OperationID: "getSecret",
		Summary:     "Get a secret",
		Method:      http.MethodGet,
		Path:        "/namespaces/{namespace}/secrets/{secret_name}",
		Tags:        []string{"secrets"},
	}, e.getSecret)

	huma.Register(api, huma.Operation{
		OperationID: "updateSecret",
		Summary:     "Update a secret",
		Method:      http.MethodPatch,
		Path:        "/namespaces/{namespace}/secrets/{secret_name}",
		Tags:        []string{"secrets"},
	}, e.updateSecret)

	huma.Register(api, huma.Operation{
		OperationID: "deleteSecret",
		Summary:     "Delete a secret",
		Method:      http.MethodDelete,
		Path:        "/namespaces/{namespace}/secrets/{secret_name}",
		Tags:        []string{"secrets"},
	}, e.deleteSecret)

	huma.Register(api, huma.Operation{
		OperationID: "createFleet",
		Summary:     "Create a fleet",
		Path:        "/fleets",
		Method:      http.MethodPost,
		Tags:        []string{"fleets"},
	}, e.createFleet)

	huma.Register(api, huma.Operation{
		OperationID: "listFleets",
		Summary:     "List fleets",
		Path:        "/fleets",
		Method:      http.MethodGet,
		Tags:        []string{"fleets"},
	}, e.listFleets)

	huma.Register(api, huma.Operation{
		OperationID: "getFleet",
		Summary:     "Get a fleet",
		Path:        "/fleets/{fleet}",
		Method:      http.MethodGet,
		Tags:        []string{"fleets"},
	}, e.getFleet)

	huma.Register(api, huma.Operation{
		OperationID: "destroyFleet",
		Summary:     "Destroy a fleet",
		Path:        "/fleets/{fleet}",
		Method:      http.MethodDelete,
		Tags:        []string{"fleets"},
	}, e.destroyFleet)

	huma.Register(api, huma.Operation{
		OperationID: "updateFleetMetadata",
		Summary:     "Update fleet metadata",
		Path:        "/fleets/{fleet}/metadata",
		Method:      http.MethodPatch,
		Tags:        []string{"fleets"},
	}, e.updateFleetMetadata)

	huma.Register(api, huma.Operation{
		OperationID: "createMachine",
		Summary:     "Create a machine",
		Method:      http.MethodPost,
		Path:        "/fleets/{fleet}/machines",
		Tags:        []string{"machines"},
	}, e.createMachine)

	huma.Register(api, huma.Operation{
		OperationID: "destroyMachine",
		Summary:     "Destroy a machine",
		Path:        "/fleets/{fleet}/machines/{machine_id}",
		Method:      http.MethodDelete,
		Tags:        []string{"machines"},
	}, e.destroyMachine)

	huma.Register(api, huma.Operation{
		OperationID: "listMachines",
		Summary:     "List machines",
		Path:        "/fleets/{fleet}/machines",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.listMachines)

	huma.Register(api, huma.Operation{
		OperationID: "getMachine",
		Summary:     "Get a machine",
		Path:        "/fleets/{fleet}/machines/{machine_id}",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.getMachine)

	huma.Register(api, huma.Operation{
		OperationID: "startMachine",
		Summary:     "Start a machine",
		Path:        "/fleets/{fleet}/machines/{machine_id}/start",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.startMachine)

	huma.Register(api, huma.Operation{
		OperationID: "stopMachine",
		Summary:     "Stop a machine",
		Path:        "/fleets/{fleet}/machines/{machine_id}/stop",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.stopMachine)

	huma.Register(api, huma.Operation{
		OperationID: "machineExec",
		Summary:     "Execute a command inside a machine",
		Path:        "/fleets/{fleet}/machines/{machine_id}/exec",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.machineExec)

	huma.Register(api, huma.Operation{
		OperationID: "listMachineVersions",
		Summary:     "List machine versions",
		Path:        "/fleets/{fleet}/machines/{machine_id}/versions",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.listMachineVersions)

	huma.Register(api, huma.Operation{
		OperationID: "listMachineEvents",
		Summary:     "List machine events",
		Method:      http.MethodGet,
		Path:        "/fleets/{fleet}/machines/{machine_id}/events",
		Tags:        []string{"machines"},
	}, e.listMachineEvents)

	huma.Register(api, huma.Operation{
		OperationID: "getMachineLogs",
		Summary:     "Get machine logs",
		Method:      http.MethodGet,
		Path:        "/fleets/{fleet}/machines/{machine_id}/logs",
		Tags:        []string{"machines"},
	}, e.getMachineLogs)

	huma.Register(api, huma.Operation{
		OperationID: "waitMachineStatus",
		Summary:     "Wait for a machine to reach a given status",
		Method:      http.MethodGet,
		Path:        "/fleets/{fleet}/machines/{machine_id}/wait",
		Tags:        []string{"machines"},
	}, e.waitMachineStatus)

	huma.Register(api, huma.Operation{
		OperationID: "enableMachineGateway",
		Summary:     "Enable the Machine Gateway",
		Method:      http.MethodPost,
		Path:        "/fleets/{fleet}/machines/{machine_id}/gateway/enable",
		Tags:        []string{"machines"},
	}, e.enableMachineGateway)

	huma.Register(api, huma.Operation{
		OperationID: "disableMachineGateway",
		Summary:     "Disable the Machine Gateway",
		Method:      http.MethodPost,
		Path:        "/fleets/{fleet}/machines/{machine_id}/gateway/disable",
		Tags:        []string{"machines"},
	}, e.disableMachineGateway)

	huma.Register(api, huma.Operation{
		OperationID: "updateMachineMetadata",
		Summary:     "Update machine metadata",
		Path:        "/fleets/{fleet}/machines/{machine_id}/metadata",
		Method:      http.MethodPatch,
		Tags:        []string{"machines"},
	}, e.updateMachineMetadata)

	huma.Register(api, huma.Operation{
		OperationID: "createGateway",
		Summary:     "Create a gateway",
		Method:      http.MethodPost,
		Path:        "/fleets/{fleet}/gateways",
		Tags:        []string{"gateways"},
	}, e.createGateway)

	huma.Register(api, huma.Operation{
		OperationID: "listGateways",
		Summary:     "List gateways",
		Method:      http.MethodGet,
		Path:        "/fleets/{fleet}/gateways",
		Tags:        []string{"gateways"},
	}, e.listGateways)

	huma.Register(api, huma.Operation{
		OperationID: "getGateway",
		Summary:     "Get a gateway",
		Method:      http.MethodGet,
		Path:        "/fleets/{fleet}/gateways/{gateway}",
		Tags:        []string{"gateways"},
	}, e.getGateway)

	huma.Register(api, huma.Operation{
		OperationID: "destroyGateway",
		Summary:     "Destroy a gateway",
		Method:      http.MethodDelete,
		Path:        "/fleets/{fleet}/gateways/{gateway}",
		Tags:        []string{"gateways"},
	}, e.destroyGateway)

	// Build endpoints
	huma.Register(api, huma.Operation{
		OperationID: "createBuild",
		Summary:     "Create a new image build",
		Method:      http.MethodPost,
		Path:        "/namespaces/{namespace}/builds",
		Tags:        []string{"builds"},
	}, e.createBuild)

	huma.Register(api, huma.Operation{
		OperationID: "listBuilds",
		Summary:     "List builds in a namespace",
		Method:      http.MethodGet,
		Path:        "/namespaces/{namespace}/builds",
		Tags:        []string{"builds"},
	}, e.listBuilds)

	huma.Register(api, huma.Operation{
		OperationID: "getBuild",
		Summary:     "Get build status",
		Method:      http.MethodGet,
		Path:        "/namespaces/{namespace}/builds/{build_id}",
		Tags:        []string{"builds"},
	}, e.getBuild)

	huma.Register(api, huma.Operation{
		OperationID: "getBuildLogs",
		Summary:     "Get build logs",
		Method:      http.MethodGet,
		Path:        "/namespaces/{namespace}/builds/{build_id}/logs",
		Tags:        []string{"builds"},
	}, e.getBuildLogs)

	huma.Register(api, huma.Operation{
		OperationID: "cancelBuild",
		Summary:     "Cancel a build",
		Method:      http.MethodPost,
		Path:        "/namespaces/{namespace}/builds/{build_id}/cancel",
		Tags:        []string{"builds"},
	}, e.cancelBuild)

	// Register health endpoints
	e.RegisterHealthEndpoints(api)
}
