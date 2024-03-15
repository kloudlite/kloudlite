package wg_vpn

import (
	"fmt"
	"net"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"github.com/miekg/dns"
	"github.com/vishvananda/netlink"
)

func getCurrentDns(verbose bool) ([]string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, err
	}

	return config.Servers, nil
}

func StartService(devName string, _ bool) error {

	// Add Wireguard device
	wgLink := &netlink.GenericLink{
		LinkAttrs: netlink.LinkAttrs{Name: devName},
		LinkType:  "wireguard",
	}
	if err := netlink.LinkAdd(wgLink); err != nil {
		return fmt.Errorf("failed to add WireGuard interface: %v", err)
	}

	// Set MTU and bring up the device
	link, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", devName, err)
	}
	if err := netlink.LinkSetMTU(link, 1420); err != nil {
		return fmt.Errorf("failed to set MTU for %s: %v", devName, err)
	}
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring up the interface %s: %v", devName, err)
	}

	return nil
}

func ipRouteAdd(ip string, _ string, devName string, _ bool) error {
	_, dst, err := net.ParseCIDR(ip)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR: %v", err)
	}

	link, err := netlink.LinkByName(devName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", devName, err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
		Scope:     netlink.SCOPE_UNIVERSE,
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("failed to add the route: %v", err)
	}

	return nil
}

func StopService(verbose bool) error {
	wgInterface, err := wgc.Show(&wgc.WgShowOptions{
		Interface: "interfaces",
	})
	if err != nil {
		return err
	}

	if len(wgInterface) == 0 {
		return nil
	}
	for _, v := range wgInterface {

		if verbose {
			fn.Log("[#] disconnecting from ", v)
		}

		if strings.TrimSpace(v) == "" {
			continue
		}

		link, err := netlink.LinkByName(strings.TrimSpace(v))
		if err != nil {
			return fmt.Errorf("failed to find the interface %s: %v", wgInterface, err)
		}

		if err := netlink.LinkDel(link); err != nil {
			return fmt.Errorf("failed to delete the interface %s: %v", wgInterface, err)
		}
	}

	return nil
}

func setDnsServer(dnsServer net.IP, deviceName string, verbose bool) error {
	return ExecCmd(fmt.Sprintf("resolvectl dns %s %s", deviceName, dnsServer.String()), verbose)
}

func SetDeviceIp(ip net.IPNet, deviceName string, _ bool) error {
	link, err := netlink.LinkByName(deviceName)
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", deviceName, err)
	}

	addr := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip.IP,
			Mask: ip.Mask,
		},
	}

	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("failed to set IP address for %s: %v", deviceName, err)
	}

	return nil
}
