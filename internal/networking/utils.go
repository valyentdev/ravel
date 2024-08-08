package networking

import (
	"net"
	"net/netip"

	"tailscale.com/net/interfaces"
)

func First(ipnet *net.IPNet) net.IP {
	first := make(net.IP, len(ipnet.IP))
	copy(first, ipnet.IP)

	for i := range ipnet.Mask {
		first[i] &= ipnet.Mask[i]
	}
	return first
}

func Last(ipnet *net.IPNet) net.IP {
	ip := ipnet.IP
	mask := ipnet.Mask
	last := make(net.IP, len(ip))
	for i := range ip {
		last[i] = ip[i] | ^mask[i]
	}
	return last
}

func NextBy(ip netip.Addr, by int) netip.Addr {
	newIp := ip
	for i := 0; i < by; i++ {
		newIp = newIp.Next()
	}
	return newIp
}

func PrevBy(ip netip.Addr, by int) netip.Addr {
	newIp := ip
	for i := 0; i < by; i++ {
		newIp = newIp.Prev()
	}
	return newIp
}

func DefaultInterface() (string, error) {
	return interfaces.DefaultRouteInterface()
}

func Overlap(a, b *net.IPNet) bool {
	return a.Contains(b.IP) || b.Contains(a.IP)
}
