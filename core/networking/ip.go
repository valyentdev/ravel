package networking

import (
	"net"
	"net/netip"
)

type IPFamily string

const (
	IPv4 IPFamily = "ipv4"
	IPv6 IPFamily = "ipv6"
)

func (f IPFamily) String() string {
	return string(f)
}

func (f IPFamily) Size() int {
	switch f {
	case IPv4:
		return 4
	case IPv6:
		return 16
	default:
		return 0
	}
}

func ParseIP(ip string) (net.IP, error) {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil, err
	}

	return addr.AsSlice(), nil
}
