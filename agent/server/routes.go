package server

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
)

func getHumaConfig() huma.Config {

	return huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:   "Ravel Agent API",
				Version: "1.0.0",
			},
		},
		OpenAPIPath: "/openapi",
		DocsPath:    "/docs",
		Formats: map[string]huma.Format{
			"application/json": huma.DefaultJSONFormat,
			"json":             huma.DefaultJSONFormat,
		},
		DefaultFormat: "application/json",
	}
}

func (s AgentServer) registerEndpoints(mux humago.Mux) {
	humaConfig := getHumaConfig()
	api := humago.New(mux, humaConfig)

	huma.Register(api, huma.Operation{
		OperationID: "putMachine",
		Path:        "/machines",
		Method:      http.MethodPost,
	}, s.putMachine)

	huma.Register(api, huma.Operation{
		OperationID: "startMachine",
		Path:        "/machines/{id}/start",
		Method:      http.MethodPost,
	}, s.startMachine)

	huma.Register(api, huma.Operation{
		OperationID: "stopMachine",
		Path:        "/machines/{id}/stop",
		Method:      http.MethodPost,
	}, s.stopMachine)

	huma.Register(api, huma.Operation{
		OperationID: "machineExec",
		Path:        "/machines/{id}/exec",
		Method:      http.MethodPost,
	}, s.machineExec)

	huma.Register(api, huma.Operation{
		OperationID: "getMachineLogs",
		Path:        "/machines/{id}/logs",
		Method:      http.MethodGet,
	}, s.getMachineLogs)

	huma.Register(api, huma.Operation{
		OperationID: "followMachineLogs",
		Path:        "/machines/{id}/logs/follow",
		Method:      http.MethodGet,
	}, s.followMachineLogs)

	huma.Register(api, huma.Operation{
		OperationID: "destroyMachine",
		Path:        "/machines/{id}",
		Method:      http.MethodDelete,
	}, s.destroyMachine)

	huma.Register(api, huma.Operation{
		OperationID: "enableMachineGateway",
		Path:        "/machines/{id}/gateway/enable",
		Method:      http.MethodPost,
	}, s.enableMachineGateway)

	huma.Register(api, huma.Operation{
		OperationID: "disableMachineGateway",
		Path:        "/machines/{id}/gateway/disable",
		Method:      http.MethodPost,
	}, s.disableMachineGateway)

	huma.Register(api, huma.Operation{
		OperationID: "waitForMachineStatus",
		Path:        "/machines/{id}/wait",
		Method:      http.MethodGet,
	}, s.waitForMachineStatus)

}
