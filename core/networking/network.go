package networking

import (
	"encoding/json"
	"errors"
	"net"
	"net/netip"
)

type networkJSON struct {
	Family       IPFamily `json:"family"`
	IP           string   `json:"ip"`
	PrefixLength int      `json:"prefix_length"`
}

type Network struct {
	Family       IPFamily `json:"family"`
	IP           net.IP   `json:"ip"`
	PrefixLength int      `json:"prefix_length"`
}

func (n *Network) Contains(ip net.IP) bool {
	return n.IPNet().Contains(ip)
}

func (n *Network) ContainsNetwork(other *Network) bool {
	return n.Contains(other.IP) && n.PrefixLength <= other.PrefixLength
}

func (n *Network) IPNet() *net.IPNet {
	if n.Family == IPv4 {
		return &net.IPNet{
			IP:   n.IP,
			Mask: net.CIDRMask(n.PrefixLength, 32),
		}
	}

	return &net.IPNet{
		IP:   n.IP,
		Mask: net.CIDRMask(n.PrefixLength, 128),
	}
}

var (
	ErrIncompatibleIPFamily = errors.New("incompatible IP family")
	ErrInvalidFamily        = errors.New("invalid family")
	ErrInvalidIP            = errors.New("invalid IP")
	ErrIncompatiblePrefix   = errors.New("incompatible prefix")
)

var _ json.Marshaler = (*Network)(nil)
var _ json.Unmarshaler = (*Network)(nil)

func (n *Network) FirstAddress() netip.Addr {
	addr, _ := netip.AddrFromSlice(First(n.IPNet()))
	return addr
}

func (n *Network) LastAddress() netip.Addr {
	addr, _ := netip.AddrFromSlice(Last(n.IPNet()))
	return addr
}

func (n *Network) MarshalJSON() ([]byte, error) {
	return json.Marshal(networkJSON{
		Family:       n.Family,
		IP:           n.IP.String(),
		PrefixLength: n.PrefixLength,
	})
}

func (n *Network) UnmarshalJSON(data []byte) error {
	var j networkJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}

	if j.Family != IPv4 && j.Family != IPv6 {
		return ErrInvalidFamily
	}

	ip, err := netip.ParseAddr(j.IP)
	if err != nil {
		return ErrInvalidIP
	}

	n.IP = ip.AsSlice()
	n.Family = j.Family
	n.PrefixLength = j.PrefixLength

	return nil
}

func (n *Network) Validate() error {
	var expectedSize int

	switch n.Family {
	case IPv4:
		expectedSize = 4
	case IPv6:
		expectedSize = 16
	default:
		return ErrInvalidFamily
	}

	if len(n.IP) != expectedSize {
		return ErrIncompatibleIPFamily
	}

	if n.PrefixLength <= 0 || n.PrefixLength > expectedSize*8 {
		return ErrIncompatiblePrefix
	}

	ipAddr, ok := netip.AddrFromSlice(n.IP)
	if !ok {
		return ErrInvalidIP
	}

	first := First(n.IPNet())

	firstAddr, ok := netip.AddrFromSlice(first)
	if !ok {
		return ErrInvalidIP
	}

	if !(ipAddr.Compare(firstAddr) == 0) {
		return ErrInvalidIP
	}

	return nil
}
