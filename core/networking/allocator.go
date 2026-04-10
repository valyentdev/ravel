package networking

import (
	"fmt"
	"net/netip"
	"sync"
)

type SubnetAllocator interface {
	Allocate(*Network) error
	Release(*Network) error
	AllocateNext() (Network, error)
}

type BasicSubnetAllocator struct {
	lock            sync.Mutex
	ipOnes          int // 32 or 128
	pool            SubnetPool
	subnetPrefixLen int
	addressBySubnet int
	firstAddress    netip.Addr
	lastAddress     netip.Addr
	used            map[netip.Addr]struct{}
}

var _ SubnetAllocator = (*BasicSubnetAllocator)(nil)

type SubnetPool struct {
	Network      Network
	SubnetPrefix int
}

func (p SubnetPool) AddressesBySubnet() int {
	bits := p.Network.Family.Size() * 8
	return 1 << (bits - p.SubnetPrefix)
}

func NewBasicSubnetAllocator(pool SubnetPool) (*BasicSubnetAllocator, error) {
	if err := pool.Network.Validate(); err != nil {
		return nil, err
	}

	addressBySubnet := pool.AddressesBySubnet()

	return &BasicSubnetAllocator{
		lock:            sync.Mutex{},
		ipOnes:          pool.Network.Family.Size(),
		pool:            pool,
		subnetPrefixLen: pool.SubnetPrefix,
		addressBySubnet: addressBySubnet,
		firstAddress:    pool.Network.FirstAddress(),
		lastAddress:     pool.Network.LastAddress(),
		used:            make(map[netip.Addr]struct{}),
	}, nil
}

func (a *BasicSubnetAllocator) Allocate(network *Network) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	if network.Family != a.pool.Network.Family {
		return fmt.Errorf("incompatible IP family")
	}

	if network.PrefixLength != a.subnetPrefixLen {
		return fmt.Errorf("incompatible prefix length: %d & %d", network.PrefixLength, a.subnetPrefixLen)
	}

	if err := network.Validate(); err != nil {
		return err
	}

	if !a.pool.Network.ContainsNetwork(network) {
		return fmt.Errorf("invalid network")
	}

	subnetAddr, ok := netip.AddrFromSlice(network.IP)
	if !ok {
		return fmt.Errorf("invalid subnet address")
	}

	a.used[subnetAddr] = struct{}{}
	return nil
}

func (a *BasicSubnetAllocator) AllocateNext() (Network, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	for ip := NextBy(a.firstAddress, a.addressBySubnet); ip.Less(PrevBy(a.lastAddress, a.addressBySubnet)); ip = NextBy(ip, a.addressBySubnet) {
		if _, ok := a.used[ip]; !ok {
			a.used[ip] = struct{}{}
			return Network{
				Family:       a.pool.Network.Family,
				IP:           ip.AsSlice(),
				PrefixLength: a.subnetPrefixLen,
			}, nil
		}
	}

	return Network{}, fmt.Errorf("no available subnet")
}

func (a *BasicSubnetAllocator) Release(subnet *Network) error {
	subnetAddr := subnet.FirstAddress()
	a.lock.Lock()
	delete(a.used, subnetAddr)
	a.lock.Unlock()
	return nil
}
