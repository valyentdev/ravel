package core

type Gateway struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	FleetId    string `json:"fleet_id"`
	Protocol   string `json:"protocol"`
	TargetPort int    `json:"target_port"`
}

type Service struct {
	Id               string `json:"id"`
	InstanceId       string `json:"instance_id"`
	Name             string `json:"name"`
	FleetId          string `json:"fleet_id"`
	LocalIPV4Address string `json:"local_ipv4_address"`
	LocalPort        int    `json:"local_port"`
	HostAddress      string `json:"host_address"`
	HostPort         int    `json:"host_port"`
}
