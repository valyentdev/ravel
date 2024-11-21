package instance

import (
	"encoding/json"
	"net"

	"github.com/valyentdev/ravel/core/networking"
)

type NetworkingConfig struct {
	TapDevice      string          `json:"tap_device"`
	Local          InstanceNetwork `json:"local"`
	DefaultGateway net.IP          `json:"default_gateway"`
}

func GetLocalNetwork(netw networking.Network) InstanceNetwork {
	first := netw.FirstAddress()
	host := networking.NextBy(first, 1).AsSlice()
	instance := networking.NextBy(first, 2).AsSlice()
	broadcast := netw.LastAddress().AsSlice()
	return InstanceNetwork{
		Network:    netw,
		HostIP:     host,
		InstanceIP: instance,
		Gateway:    host,
		Broadcast:  broadcast,
	}
}

type networkingConfigJSON struct {
	TapDevice      string          `json:"tap_device"`
	Local          InstanceNetwork `json:"local"`
	DefaultGateway string          `json:"default_gateway"`
}

var _ json.Marshaler = (*NetworkingConfig)(nil)
var _ json.Unmarshaler = (*NetworkingConfig)(nil)

// MarshalJSON implements json.Marshaler.
func (n *NetworkingConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(networkingConfigJSON{
		TapDevice:      n.TapDevice,
		Local:          n.Local,
		DefaultGateway: n.DefaultGateway.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NetworkingConfig) UnmarshalJSON(bytes []byte) error {
	var j networkingConfigJSON
	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	defaultGateway, err := networking.ParseIP(j.DefaultGateway)
	if err != nil {
		return err
	}

	n.Local = j.Local
	n.DefaultGateway = defaultGateway
	n.TapDevice = j.TapDevice

	return nil
}

type InstanceNetwork struct {
	Network    networking.Network `json:"network"`
	InstanceIP net.IP             `json:"instance_ip"`
	HostIP     net.IP             `json:"host_ip"`
	Gateway    net.IP             `json:"gateway"`
	Broadcast  net.IP             `json:"broadcast"`
}

func (i *InstanceNetwork) HostIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   i.HostIP,
		Mask: i.Network.IPNet().Mask,
	}
}

func (i *InstanceNetwork) InstanceIPNet() *net.IPNet {

	return &net.IPNet{
		IP:   i.InstanceIP,
		Mask: i.Network.IPNet().Mask,
	}
}

var _ json.Marshaler = (*InstanceNetwork)(nil)
var _ json.Unmarshaler = (*InstanceNetwork)(nil)

type instanceNetworkJSON struct {
	Network    networking.Network `json:"network"`
	InstanceIp string             `json:"instance_ip"`
	HostIp     string             `json:"host_ip"`
	Gateway    string             `json:"gateway"`
	Broadcast  string             `json:"broadcast"`
}

// MarshalJSON implements json.Marshaler.
func (i *InstanceNetwork) MarshalJSON() ([]byte, error) {
	return json.Marshal(instanceNetworkJSON{
		Network:    i.Network,
		InstanceIp: i.InstanceIP.String(),
		HostIp:     i.HostIP.String(),
		Gateway:    i.Gateway.String(),
		Broadcast:  i.Broadcast.String(),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (i *InstanceNetwork) UnmarshalJSON(bytes []byte) error {
	var j instanceNetworkJSON
	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	instanceIP, err := networking.ParseIP(j.InstanceIp)
	if err != nil {
		return err
	}

	hostIP, err := networking.ParseIP(j.HostIp)
	if err != nil {
		return err
	}

	gateway, err := networking.ParseIP(j.Gateway)
	if err != nil {
		return err
	}

	broadcast, err := networking.ParseIP(j.Broadcast)
	if err != nil {
		return err
	}

	i.Network = j.Network
	i.InstanceIP = instanceIP
	i.HostIP = hostIP
	i.Gateway = gateway
	i.Broadcast = broadcast

	return nil
}
