package tap

import (
	"errors"

	"github.com/vishvananda/netlink"
)

func TapName(name string) string {
	if len(name) > 15 {
		return name[:15]
	}

	return name
}

func createTap(name string) error {
	tap := &netlink.Tuntap{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
		},
		Mode: netlink.TUNTAP_MODE_TAP,
	}
	if err := netlink.LinkAdd(tap); err != nil {
		return err
	}

	if err := netlink.LinkSetUp(tap); err != nil {
		return err
	}

	return nil
}

func deleteTap(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		if errors.Is(err, netlink.LinkNotFoundError{}) {
			return nil
		}
		return err
	}

	if err := netlink.LinkDel(link); err != nil {
		return err
	}

	return nil
}
