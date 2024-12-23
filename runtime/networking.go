package runtime

import (
	"net"

	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/networking"
	"github.com/valyentdev/ravel/core/networking/tap"
	"github.com/valyentdev/ravel/internal/id"
	"github.com/valyentdev/ravel/runtime/vm"
)

type networkService struct {
	localSubnetAllocator *networking.BasicSubnetAllocator
	jailerUser           vm.User
}

var _ instance.NetworkingService = (*networkService)(nil)

func newNetworkService(user vm.User) *networkService {
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

	return &networkService{
		localSubnetAllocator: localSubnetAllocator,
		jailerUser:           user,
	}
}

func (n *networkService) CleanupInstanceNetwork(id string, config instance.NetworkingConfig) error {
	return tap.CleanupInstanceTapDevice(id, config)
}

func (n *networkService) EnsureInstanceNetwork(id string, config instance.NetworkingConfig) error {
	_, err := tap.PrepareInstanceTapDevice(id, config, n.jailerUser.Uid, n.jailerUser.Gid)
	return err
}

func (n *networkService) Allocate(in instance.NetworkingConfig) error {
	return n.localSubnetAllocator.Allocate(&in.Local.Network)
}

func (n *networkService) AllocateNext() (instance.NetworkingConfig, error) {
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

func (n *networkService) Release(network instance.NetworkingConfig) {
	n.localSubnetAllocator.Release(&network.Local.Network)
}
