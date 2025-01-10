package api

type CreateGatewayPayload struct {
	Name       string `json:"name,omitempty"`
	Fleet      string `json:"fleet"`
	TargetPort int    `json:"target_port"`
}

type Gateway struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	FleetId    string `json:"fleet_id"`
	Protocol   string `json:"protocol"`
	TargetPort int    `json:"target_port"`
}
