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

func (s DaemonServer) registerEndpoints(mux humago.Mux) {
	humaConfig := getHumaConfig()
	api := humago.New(mux, humaConfig)

	huma.Register(api, huma.Operation{
		OperationID: "createInstance",
		Path:        "/instances",
		Method:      http.MethodPost,
	}, s.createInstance)

	huma.Register(api, huma.Operation{
		OperationID: "listInstances",
		Path:        "/instances",
		Method:      http.MethodGet,
	}, s.listInstances)

	huma.Register(api, huma.Operation{
		OperationID: "getInstance",
		Path:        "/instances/{id}",
		Method:      http.MethodGet,
	}, s.getInstance)

	huma.Register(api, huma.Operation{
		OperationID: "destroyInstance",
		Path:        "/instances/{id}",
		Method:      http.MethodDelete,
	}, s.destroyInstance)

	huma.Register(api, huma.Operation{
		OperationID: "startInstance",
		Path:        "/instances/{id}/start",
		Method:      http.MethodPost,
	}, s.startInstance)

	huma.Register(api, huma.Operation{
		OperationID: "stopInstance",
		Path:        "/instances/{id}/stop",
		Method:      http.MethodPost,
	}, s.stopInstance)

	huma.Register(api, huma.Operation{
		OperationID: "exec",
		Path:        "/instances/{id}/exec",
		Method:      http.MethodPost,
	}, s.exec)

	huma.Register(api, huma.Operation{
		OperationID: "getInstanceLogs",
		Path:        "/instances/{id}/logs",
		Method:      http.MethodGet,
	}, s.getInstanceLogs)

	huma.Register(api, huma.Operation{
		OperationID: "followInstanceLogs",
		Path:        "/instances/{id}/logs/follow",
		Method:      http.MethodGet,
	}, s.followInstanceLogs)

	huma.Register(api, huma.Operation{
		OperationID: "listImages",
		Path:        "/images",
		Method:      http.MethodGet,
	}, s.listImages)

	huma.Register(api, huma.Operation{
		OperationID: "pullImage",
		Path:        "/images/pull",
		Method:      http.MethodPost,
	}, s.pullImage)

	huma.Register(api, huma.Operation{
		OperationID: "deleteImage",
		Path:        "/images/{ref}",
		Method:      http.MethodDelete,
	}, s.deleteImage)

}
