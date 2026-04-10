package endpoints

type NSResolver struct {
	Namespace string `query:"namespace" required:"true"`
}

type FleetResolver struct {
	NSResolver
	Fleet string `path:"fleet"`
}

type MachineResolver struct {
	FleetResolver
	MachineId string `path:"machine_id"`
}
