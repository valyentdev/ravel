package network

import (
	"net"

	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/networking"
	"github.com/valyentdev/ravel/internal/id"
)

type NetworkService struct {
	localSubnetAllocator *networking.BasicSubnetAllocator
}

func NewNetworkService() *NetworkService {
	localSubnetAllocator, err := networking.NewBasicSubnetAllocator(networking.SubnetPool{
		Network: networking.Network{
			Family:       networking.IPv4,
			IP:           net.IPv4(172, 18, 0, 0).To4(),
			PrefixLength: 16,
		},
		SubnetPrefix: 29,
	})

	if err != nil {
		panic(err)
	}

	return &NetworkService{
		localSubnetAllocator: localSubnetAllocator,
	}
}

func (n *NetworkService) Allocate(in instance.NetworkingConfig) error {
	return n.localSubnetAllocator.Allocate(&in.Local.Network)
}

func (n *NetworkService) AllocateNext() (instance.NetworkingConfig, error) {
	net, err := n.localSubnetAllocator.AllocateNext()
	if err != nil {
		return instance.NetworkingConfig{}, err
	}

	local := instance.GetLocalNetwork(net)

	return instance.NetworkingConfig{
		TapDevice:      id.Generate()[:14],
		Local:          local,
		DefaultGateway: local.HostIP,
	}, nil
}

func (n *NetworkService) Release(network instance.NetworkingConfig) {
	n.localSubnetAllocator.Release(&network.Local.Network)
}
