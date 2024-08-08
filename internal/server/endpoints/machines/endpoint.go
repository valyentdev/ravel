package machines

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/internal/server/utils"
	"github.com/valyentdev/ravel/pkg/manager"
)

type Endpoint struct {
	validator *utils.Validator
	m         *manager.Manager
}

func NewEndpoint(m *manager.Manager, validator *utils.Validator) *Endpoint {
	return &Endpoint{
		validator: validator,
		m:         m,
	}
}

func (e *Endpoint) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createMachine",
		Summary:     "Create a machine",
		Method:      http.MethodPost,
		Path:        "/fleets/{fleet_id}/machines",
		Tags:        []string{"machines"},
	}, e.createMachine)

	huma.Register(api, huma.Operation{
		OperationID: "destroyMachine",
		Summary:     "Destroy a machine",
		Path:        "/fleets/{fleet_id}/machines/{machine_id}",
		Method:      http.MethodDelete,
		Tags:        []string{"machines"},
	}, e.destroyMachine)

	huma.Register(api, huma.Operation{
		OperationID: "listMachines",
		Summary:     "List machines",
		Path:        "/fleets/{fleet_id}/machines",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.listMachines)

	huma.Register(api, huma.Operation{
		OperationID: "getMachine",
		Summary:     "Get a machine",
		Path:        "/fleets/{fleet_id}/machines/{machine_id}",
		Method:      http.MethodGet,
		Tags:        []string{"machines"},
	}, e.getMachine)

	huma.Register(api, huma.Operation{
		OperationID: "startMachine",
		Summary:     "Start a machine",
		Path:        "/fleets/{fleet_id}/machines/{machine_id}/start",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.startMachine)

	huma.Register(api, huma.Operation{
		OperationID: "stopMachine",
		Summary:     "Stop a machine",
		Path:        "/fleets/{fleet_id}/machines/{machine_id}/stop",
		Method:      http.MethodPost,
		Tags:        []string{"machines"},
	}, e.stopMachine)
}
