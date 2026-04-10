package api

import "time"

type Secret struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSecretPayload struct {
	Name  string `json:"name" minLength:"1" maxLength:"63" pattern:"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$" doc:"Secret name (DNS label format)"`
	Value string `json:"value" minLength:"1" doc:"Secret value"`
}

type UpdateSecretPayload struct {
	Value string `json:"value" minLength:"1" doc:"New secret value"`
}
