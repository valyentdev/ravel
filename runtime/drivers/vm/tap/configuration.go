package tap

import (
	"errors"
	"fmt"

	"github.com/coreos/go-iptables/iptables"
	"github.com/valyentdev/ravel/core/instance"
	"github.com/valyentdev/ravel/core/networking"
	"github.com/vishvananda/netlink"
)

func configureTapDevice(tap string, config instance.NetworkingConfig) error {
	link, err := netlink.LinkByName(tap)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{
		IPNet: config.Local.HostIPNet(),
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

	network := config.Local.Network.IPNet().String()

	if err := ipt.AppendUnique("filter", "FORWARD", "-i", tap, "-o", defaultInterface, "-j", "ACCEPT"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("nat", "POSTROUTING", "-s", network, "-o", defaultInterface, "-j", "MASQUERADE"); err != nil {
		return err
	}
	if err := ipt.AppendUnique("filter", "FORWARD", "-s", network, "-o", defaultInterface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		return err
	}

	return nil
}

func cleanupTapDeviceConfig(config instance.NetworkingConfig) error {
	tap := config.TapDevice
	errs := []error{}

	ipt, err := iptables.New()
	if err != nil {
		return fmt.Errorf("failed to create iptables client: %w", err)
	}

	defaultInterface, err := networking.DefaultInterface()
	if err != nil {
		return fmt.Errorf("failed to get default interface: %w", err)
	}

	network := config.Local.Network.IPNet().String()

	if err := ipt.DeleteIfExists("filter", "FORWARD", "-i", tap, "-o", defaultInterface, "-j", "ACCEPT"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	if err := ipt.DeleteIfExists("nat", "POSTROUTING", "-s", network, "-o", defaultInterface, "-j", "MASQUERADE"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	if err := ipt.DeleteIfExists("filter", "FORWARD", "-s", network, "-o", defaultInterface, "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT"); err != nil {
		errs = append(errs, fmt.Errorf("failed to delete iptables rule for tap device: %w", err))
	}

	return errors.Join(errs...)
}
