package core

import "time"

type Node struct {
	Id            string    `json:"id"`
	Address       string    `json:"address"`
	Region        string    `json:"region"`
	HeartbeatedAt time.Time `json:"heartbeated_at"`
}
