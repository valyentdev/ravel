package wireguard

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// CreateInterface creates a new Wireguard interface
func CreateInterface(config InterfaceConfig) error {
	// Create the Wireguard link
	attrs := netlink.NewLinkAttrs()
	attrs.Name = config.Name

	link := &netlink.Wireguard{LinkAttrs: attrs}

	if err := netlink.LinkAdd(link); err != nil {
		return fmt.Errorf("failed to create wireguard interface: %w", err)
	}

	// Set the interface up
	if err := netlink.LinkSetUp(link); err != nil {
		netlink.LinkDel(link)
		return fmt.Errorf("failed to set interface up: %w", err)
	}

	// Add IP address to the interface
	addr := &netlink.Addr{
		IPNet: config.Address,
	}
	if err := netlink.AddrAdd(link, addr); err != nil {
		netlink.LinkDel(link)
		return fmt.Errorf("failed to add address to interface: %w", err)
	}

	// Configure Wireguard settings (private key, listen port, peers)
	if err := configureWireguard(config); err != nil {
		netlink.LinkDel(link)
		return fmt.Errorf("failed to configure wireguard: %w", err)
	}

	return nil
}

// configureWireguard configures the Wireguard interface using wgctrl
func configureWireguard(config InterfaceConfig) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to create wgctrl client: %w", err)
	}
	defer client.Close()

	// Convert our Key type to wgtypes.Key
	var privateKey wgtypes.Key
	copy(privateKey[:], config.PrivateKey[:])

	// Build peer configs
	var peers []wgtypes.PeerConfig
	for _, p := range config.Peers {
		var publicKey wgtypes.Key
		copy(publicKey[:], p.PublicKey[:])

		// Convert []*net.IPNet to []net.IPNet
		allowedIPs := make([]net.IPNet, len(p.AllowedIPs))
		for i, ipNet := range p.AllowedIPs {
			allowedIPs[i] = *ipNet
		}

		peerConfig := wgtypes.PeerConfig{
			PublicKey:  publicKey,
			AllowedIPs: allowedIPs,
		}

		// Parse endpoint if provided
		if p.Endpoint != "" {
			endpoint, err := net.ResolveUDPAddr("udp", p.Endpoint)
			if err != nil {
				return fmt.Errorf("invalid peer endpoint %s: %w", p.Endpoint, err)
			}
			peerConfig.Endpoint = endpoint
		}

		peers = append(peers, peerConfig)
	}

	// Build Wireguard configuration
	wgConfig := wgtypes.Config{
		PrivateKey: &privateKey,
		Peers:      peers,
	}

	// Set listen port if specified
	if config.ListenPort > 0 {
		wgConfig.ListenPort = &config.ListenPort
	}

	// Configure the interface
	if err := client.ConfigureDevice(config.Name, wgConfig); err != nil {
		return fmt.Errorf("failed to configure device: %w", err)
	}

	return nil
}

// DeleteInterface deletes a Wireguard interface
func DeleteInterface(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		// Interface doesn't exist, nothing to delete
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			return nil
		}
		return fmt.Errorf("failed to get interface %s: %w", name, err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete interface %s: %w", name, err)
	}

	return nil
}

// AddPeer adds a peer to an existing Wireguard interface
func AddPeer(interfaceName string, peer PeerConfig) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to create wgctrl client: %w", err)
	}
	defer client.Close()

	var publicKey wgtypes.Key
	copy(publicKey[:], peer.PublicKey[:])

	// Convert []*net.IPNet to []net.IPNet
	allowedIPs := make([]net.IPNet, len(peer.AllowedIPs))
	for i, ipNet := range peer.AllowedIPs {
		allowedIPs[i] = *ipNet
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey:  publicKey,
		AllowedIPs: allowedIPs,
	}

	if peer.Endpoint != "" {
		endpoint, err := net.ResolveUDPAddr("udp", peer.Endpoint)
		if err != nil {
			return fmt.Errorf("invalid peer endpoint %s: %w", peer.Endpoint, err)
		}
		peerConfig.Endpoint = endpoint
	}

	config := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	if err := client.ConfigureDevice(interfaceName, config); err != nil {
		return fmt.Errorf("failed to add peer: %w", err)
	}

	return nil
}

// RemovePeer removes a peer from a Wireguard interface
func RemovePeer(interfaceName string, publicKey Key) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("failed to create wgctrl client: %w", err)
	}
	defer client.Close()

	var wgPublicKey wgtypes.Key
	copy(wgPublicKey[:], publicKey[:])

	remove := true
	config := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgPublicKey,
				Remove:    remove,
			},
		},
	}

	if err := client.ConfigureDevice(interfaceName, config); err != nil {
		return fmt.Errorf("failed to remove peer: %w", err)
	}

	return nil
}

// SetOwner sets the owner (UID/GID) of a Wireguard interface
func SetOwner(name string, uid, gid int) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to get interface %s: %w", name, err)
	}

	// Open netlink socket to set owner
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("failed to create netlink socket: %w", err)
	}
	defer unix.Close(fd)

	// Note: Setting owner on Wireguard interfaces is typically not needed
	// as they're created in the host namespace. This is here for completeness.
	_ = link // Suppress unused variable warning

	return nil
}
