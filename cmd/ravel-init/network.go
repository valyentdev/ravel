package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/valyentdev/ravel/internal/vminit"
	"github.com/vishvananda/netlink"
)

func setupLoopback() error {
	lo, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("error getting loopback interface: %v", err)
	}

	if err := netlink.LinkSetUp(lo); err != nil {
		return fmt.Errorf("error configuring loopback interface: %v", err)
	}
	return nil
}

func getNetlinkAddr(ipConfig vminit.IPConfig) (*netlink.Addr, error) {
	ip, ipNet, err := net.ParseCIDR(ipConfig.IPNet)
	if err != nil {
		return nil, fmt.Errorf("error parsing IP address: %v", err)
	}

	broadcast := net.ParseIP(ipConfig.Broadcast)

	if !ipNet.Contains(broadcast) {
		return nil, fmt.Errorf("broadcast address %s is not in the network %s", broadcast, ipNet)
	}

	if len(ipNet.IP) == 4 {
		ip = ip.To4()
		broadcast = broadcast.To4()
	} else if len(ipNet.IP) == 16 {
		ip = ip.To16()
		broadcast = broadcast.To16()
	}

	mask := ipNet.Mask

	return &netlink.Addr{
		Broadcast: broadcast,
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: net.IPMask(mask),
		},
	}, nil
}

func ApplyIPConfig(link netlink.Link, ipConfig vminit.IPConfig) error {
	addr, err := getNetlinkAddr(ipConfig)
	if err != nil {
		return fmt.Errorf("error getting netlink address: %v", err)
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("error adding IP address to interface: %v", err)
	}

	return nil
}

func NetworkSetup(config vminit.NetworkConfig) error {
	slog.Debug("setting network interfaces up")

	if err := setupLoopback(); err != nil {
		return fmt.Errorf("error setting up loopback interface: %v", err)
	}

	eth0, err := netlink.LinkByName("eth0")
	if err != nil {
		return fmt.Errorf("error getting eth0 interface: %v", err)
	}

	for _, v := range config.IPConfigs {
		err := ApplyIPConfig(eth0, v)
		if err != nil {
			return fmt.Errorf("error applying IP config: %v", err)
		}
	}

	slog.Debug("Setting eth0 interface up", "eth0", eth0)
	if err := netlink.LinkSetUp(eth0); err != nil {
		return fmt.Errorf("error setting eth0 interface up: %v", err)
	}

	slog.Debug("Adding default route", "gateway", config.DefaultGateway)
	if err := netlink.RouteAdd(&netlink.Route{
		Gw: net.ParseIP(config.DefaultGateway),
	}); err != nil {
		return fmt.Errorf("error adding default route: %v", err)
	}

	return nil
}

func WriteEtcResolv(entries vminit.EtcResolv) error {
	slog.Debug("populating /etc/resolv.conf")

	f, err := os.OpenFile("/etc/resolv.conf", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error opening resolv.conf: %v", err)
	}
	defer f.Close()

	for _, entry := range entries.Nameservers {
		if _, err := fmt.Fprintf(f, "\nnameserver\t%s", entry); err != nil {
			return fmt.Errorf("error writing to resolv.conf file: %v", err)
		}
	}

	if _, err := f.Write([]byte("\n")); err != nil {
		return err
	}

	return nil
}

var (
	defaultHosts = []vminit.EtcHost{
		{IP: "127.0.0.1", Host: "localhost localhost.localdomain localhost4 localhost4.localdomain4"},
		{IP: "::1", Host: "localhost localhost.localdomain localhost6 localhost6.localdomain6"},
		{IP: "fe00::0", Host: "ip6-localnet"},
		{IP: "ff00::0", Host: "ip6-mcastprefix"},
		{IP: "ff02::1", Host: "ip6-allnodes"},
		{IP: "ff02::2", Host: "ip6-allrouters"},
	}

	etchostPath = "/etc/hosts"
)

func WriteEtcHost(hosts []vminit.EtcHost) error {
	slog.Debug("populating /etc/hosts")

	records := append(defaultHosts, hosts...)

	f, err := os.OpenFile(etchostPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("error opening /etc/hosts file: %v", err)
	}
	defer f.Close()

	for _, entry := range records {
		if entry.Desc != "" {
			_, err := fmt.Fprintf(f, "# %s\n%s\t%s\n", entry.Desc, entry.IP, entry.Host)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintf(f, "%s\t%s\n", entry.IP, entry.Host)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
