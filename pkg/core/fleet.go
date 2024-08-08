package core

import "time"

type Fleet struct {
	Id        string    `json:"id"`
	Namespace string    `json:"namespace"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
