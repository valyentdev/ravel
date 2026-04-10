package wireguard

import (
	"net"
)

// InterfaceConfig represents the configuration for a Wireguard interface
type InterfaceConfig struct {
	Name       string       // Interface name (e.g., "wg0")
	PrivateKey Key          // Private key for this interface
	Address    *net.IPNet   // IP address and network for this interface
	ListenPort int          // UDP port to listen on (0 for automatic)
	Peers      []PeerConfig // List of peers
}

// PeerConfig represents a Wireguard peer configuration
type PeerConfig struct {
	PublicKey  Key          // Public key of the peer
	AllowedIPs []*net.IPNet // IP ranges allowed from this peer
	Endpoint   string       // Endpoint address (e.g., "1.2.3.4:51820"), optional
}

// NetworkConfig represents a private network configuration
type NetworkConfig struct {
	Name          string       // Network name
	NetworkCIDR   string       // Network CIDR (e.g., "10.0.1.0/24")
	InterfaceName string       // Wireguard interface name
	PrivateKey    Key          // Private key for this machine
	PublicKey     Key          // Public key for this machine
	IPAddress     *net.IPNet   // IP address for this machine in the network
	Peers         []PeerConfig // Peer configurations
}
