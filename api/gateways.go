package api

type CreateGatewayPayload struct {
	Name       string `json:"name"`
	Fleet      string `json:"fleet"`
	TargetPort int    `json:"target_port"`
}
