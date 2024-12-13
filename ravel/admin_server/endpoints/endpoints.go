package endpoints

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/ravel"
)

// Endpoints wraps the administration server-related endpionts.
type Endpoints struct {
	ravel *ravel.Ravel
}

// New creates a new instance of the endpoints wrapper.
func New() *Endpoints {
	return &Endpoints{}
}

// Register maps the endpoints handlers functions to actual Huma REST endpoints.
func (e *Endpoints) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "listNodes",
		Path:        "/nodes",
		Method:      http.MethodGet,
		Tags:        []string{"nodes"},
		Summary:     "List nodes",
	}, e.listNodes)

	e.registerNamespacesEndpoints(api)
}

func (e *Endpoints) registerNamespacesEndpoints(api huma.API) {
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
}
