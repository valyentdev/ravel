package networking

import (
	"net"
	"net/netip"

	"github.com/valyentdev/ravel/internal/vminit"
)

type LocalIPV4Subnet string

func (l LocalIPV4Subnet) LocalConfig() LocalConfig {
	_, ipnet, _ := net.ParseCIDR(string(l))

	subnetAddr, _ := netip.AddrFromSlice(ipnet.IP)

	hostIp := NextBy(subnetAddr, 1)
	machineIp := NextBy(subnetAddr, 2)

	return LocalConfig{
		Network:   ipnet,
		HostIP:    hostIp.AsSlice(),
		MachineIP: machineIp.AsSlice(),
	}
}

type LocalConfig struct {
	Network   *net.IPNet
	HostIP    net.IP
	MachineIP net.IP
}

func (l LocalConfig) InitConfig() vminit.IPConfig {
	ipnet := net.IPNet{
		IP:   l.MachineIP,
		Mask: l.Network.Mask,
	}
	broadcast := Last(l.Network)
	return vminit.IPConfig{
		IPNet:     ipnet.String(),
		Gateway:   l.HostIP.String(),
		Broadcast: broadcast.String(),
	}
}
