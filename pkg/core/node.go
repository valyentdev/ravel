package core

import (
	"fmt"
	"time"
)

type Node struct {
	Id            string    `json:"id"`
	Address       string    `json:"address"`
	AgentPort     int       `json:"agent_port"`
	HttpProxyPort int       `json:"proxy_port"`
	Region        string    `json:"region"`
	HeartbeatedAt time.Time `json:"heartbeated_at"`
}

func (n *Node) AgentAddress() string {
	return fmt.Sprintf("%s:%d", n.Address, n.AgentPort)
}

func (n *Node) HttpProxyAddress() string {
	return fmt.Sprintf("%s:%d", n.Address, n.HttpProxyPort)
}
