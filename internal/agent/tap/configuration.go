package tap

import (
	"errors"
	"fmt"
	"net"

	"github.com/coreos/go-iptables/iptables"
	"github.com/valyentdev/ravel/internal/networking"
	"github.com/valyentdev/ravel/pkg/core"
	"github.com/vishvananda/netlink"
)

func configureTapDevice(tap string, machine core.Instance) error {
	config := networking.LocalConfig{}
	link, err := netlink.LinkByName(tap)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   config.HostIP,
			Mask: config.Network.Mask,
		},
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		return err
	}

	ipt, err := iptables.New()
	if err != nil {
		return err
	}

	defaultInterface, err := networking.DefaultInterface()
	if err != nil {
		return err
	}

	if err := ipt.AppendUnique("filter", "FORWARD", "-i", tap, "-o", defaultInterface, "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("nat", "POSTROUTING", "-s", config.Network.String(), "-o", defaultInterface, "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "FORWARD", "-s", config.Network.String(), "-o", defaultInterface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}

func cleanupTapDeviceConfig(tap string, machine core.Instance) error {
	errs := []error{}

	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("failed to create iptables client: %w", err)
	}

	defaultInterface, err := networking.DefaultInterface()
	if err != nil {
		return fmt.Errorf("failed to get default interface: %w", err)
	}

	if err := ipt.DeleteIfExists("filter", "FORWARD", "-i", tap, "-o", defaultInterface, "-j", "ACCEPT"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	config := networking.LocalConfig{}
	if err := ipt.DeleteIfExists("nat", "POSTROUTING", "-s", config.Network.String(), "-o", defaultInterface, "-j", "MASQUERADE"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	if err := ipt.DeleteIfExists("filter", "FORWARD", "-s", config.Network.String(), "-o", defaultInterface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	return errors.Join(errs...)
}
