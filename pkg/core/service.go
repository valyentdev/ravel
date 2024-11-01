package core

type Gateway struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	FleetId    string `json:"fleet_id"`
	Protocol   string `json:"protocol"`
	TargetPort int    `json:"target_port"`
}
