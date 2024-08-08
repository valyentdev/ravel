package nodes

import (
	"github.com/danielgtaylor/huma/v2"
	"github.com/valyentdev/ravel/pkg/manager"
)

type Endpoint struct {
	m *manager.Manager
}

func NewEndpoint(m *manager.Manager) *Endpoint {
	return &Endpoint{m: m}
}

func (e *Endpoint) Register(api huma.API) {
	huma.Get(api, "/nodes", e.listNodes)
}
