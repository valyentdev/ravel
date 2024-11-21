package endpoints

import "github.com/danielgtaylor/huma/v2"

type NSResolver struct {
	Namespace string `query:"namespace"`
}

func (r *NSResolver) Resolve(ctx huma.Context) []error {
	if r.Namespace == "" {
		r.Namespace = "default"
	}
	return nil
}

var _ huma.Resolver = &NSResolver{}

type FleetResolver struct {
	NSResolver
	Fleet string `path:"fleet"`
}

type MachineResolver struct {
	FleetResolver
	MachineId string `path:"machine_id"`
}
