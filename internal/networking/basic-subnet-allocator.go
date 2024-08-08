package networking

import (
	"fmt"
	"net"
	"net/netip"
	"sync"
)

type BasicSubnetAllocator struct {
	lock            sync.Mutex
	ipSize          int // 32 or 128
	pool            IPNetPool
	subnetPrefixLen int
	addressBySubnet int
	firstAddress    netip.Addr
	lastAddress     netip.Addr
	used            map[netip.Addr]struct{}
}

type IPNetPool struct {
	Pool       net.IPNet
	SubnetMask net.IPMask
}

func NewBasicSubnetAllocator(pool IPNetPool) (*BasicSubnetAllocator, error) {
	if pool.Pool.IP == nil || pool.Pool.Mask == nil || pool.SubnetMask == nil {
		return nil, fmt.Errorf("invalid pool")
	}

	if !(len(pool.Pool.IP) == len(pool.SubnetMask) && len(pool.Pool.IP) == len(pool.Pool.Mask)) {
		return nil, fmt.Errorf("invalid pool: different lengths")
	}

	poolOnes, _ := pool.Pool.Mask.Size()
	subnetOnes, _ := pool.SubnetMask.Size()

	if poolOnes > subnetOnes {
		return nil, fmt.Errorf("invalid pool: subnet mask is greater than pool mask")
	}

	subnetPrefixLen, _ := pool.SubnetMask.Size()

	addressBySubnet := 1 << uint(len(pool.Pool.IP)*8-subnetPrefixLen)

	firstAddress, ok := netip.AddrFromSlice(First(&pool.Pool))
	if !ok {
		return nil, fmt.Errorf("invalid pool address: must be the first address")
	}

	lastAddress, ok := netip.AddrFromSlice(Last(&pool.Pool))
	if !ok {
		return nil, fmt.Errorf("invalid pool address: cannot find the last address")
	}

	return &BasicSubnetAllocator{
		lock:            sync.Mutex{},
		ipSize:          len(pool.Pool.IP) * 8,
		pool:            pool,
		subnetPrefixLen: subnetPrefixLen,
		addressBySubnet: addressBySubnet,
		firstAddress:    firstAddress,
		lastAddress:     lastAddress,
		used:            make(map[netip.Addr]struct{}),
	}, nil
}

func (a *BasicSubnetAllocator) Allocate(ipnet *net.IPNet) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if ipnet == nil {
		return fmt.Errorf("invalid subnet")
	}

	if ipnet.IP == nil || ipnet.Mask == nil {
		return fmt.Errorf("invalid subnet")
	}

	if len(ipnet.IP) != len(a.pool.Pool.IP) || len(ipnet.Mask) != len(a.pool.Pool.Mask) {
		return fmt.Errorf("invalid subnet: different lengths")
	}

	if !a.pool.Pool.Contains(ipnet.IP) {
		return fmt.Errorf("invalid subnet: not in the pool")
	}

	ones, _ := ipnet.Mask.Size()
	if ones != a.subnetPrefixLen {
		return fmt.Errorf("invalid subnet: prefix length must be %d", a.subnetPrefixLen)
	}

	first := First(ipnet)

	subnetAddr, ok := netip.AddrFromSlice(first)
	if !ok {
		return fmt.Errorf("invalid subnet address")
	}

	a.used[subnetAddr] = struct{}{}
	return nil
}

func (a *BasicSubnetAllocator) AllocateNext() (net.IPNet, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for ip := NextBy(a.firstAddress, a.addressBySubnet); ip.Less(PrevBy(a.lastAddress, a.addressBySubnet)); ip = NextBy(ip, a.addressBySubnet) {

		subnet := net.IPNet{
			IP:   net.IP(ip.AsSlice()),
			Mask: net.CIDRMask(a.subnetPrefixLen, a.ipSize),
		}

		if _, ok := a.used[ip]; !ok {
			a.used[ip] = struct{}{}
			return subnet, nil
		}
	}
	return net.IPNet{}, fmt.Errorf("no available subnet")
}

func (a *BasicSubnetAllocator) Release(subnet *net.IPNet) error {
	if subnet == nil {
		return fmt.Errorf("invalid subnet")
	}

	a.lock.Lock()
	defer a.lock.Unlock()

	first := First(subnet)

	subnetAddr, ok := netip.AddrFromSlice(first)
	if !ok {
		return fmt.Errorf("invalid subnet address")
	}

	delete(a.used, subnetAddr)
	return nil
}
