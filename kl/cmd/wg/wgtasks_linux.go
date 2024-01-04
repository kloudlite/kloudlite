package wg

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/vishvananda/netlink"
)

func configureDarwin(_ string, _ bool) error {
	// not required to implement
	return nil
}

func connect(verbose bool) error {
	success := false
	defer func() {
		if !success {
			stopService(verbose)
		}
	}()

	startService(verbose)

	if err := startConfiguration(verbose); err != nil {
		return err
	}
	success = true
	return nil
}

func disconnect(verbose bool) error {
	return stopService(verbose)
}

func setDNS(dns []net.IP, verbose bool) error {
	if verbose {
		fn.Log("[#] setting /etc/resolv.conf")
	}

	file, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err
	}

	if _, e := os.Stat("/etc/resolv.conf.bak"); errors.Is(e, os.ErrNotExist) {
		e = os.WriteFile("/etc/resolv.conf.bak", file, 0644)
		if e != nil {
			return e
		}
	}

	dnsArr := make([]string, 0)

	err = os.WriteFile("/etc/resolv.conf", []byte(func() string {
		resolveString := ""

		for _, i2 := range dns {
			dnsArr = append(dnsArr, i2.String())
			resolveString += fmt.Sprintf("nameserver %s\n", i2.String())
		}
		resolveString += "nameserver 8.8.8.8\n"

		client.SetActiveDns(dnsArr)

		return resolveString
	}()), 0644)

	return err
}

func resetDNS(verbose bool) error {

	if verbose {
		fn.Log("[#] resetting /etc/resolv.conf")
	}

	if _, e := os.Stat("/etc/resolv.conf"); e != nil && !errors.Is(e, os.ErrNotExist) {
		if err := os.Remove("/etc/resolv.conf"); err != nil {
			return err
		}
	}

	if _, e := os.Stat("/etc/resolv.conf.bak"); errors.Is(e, os.ErrNotExist) {
		return nil
	}

	return os.Rename("/etc/resolv.conf.bak", "/etc/resolv.conf")
}

func startService(_ bool) error {
	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

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

func stopService(verbose bool) error {
	wgInterface, err := wgc.Show(&wgc.WgShowOptions{
		Interface: "interfaces",
	})
	if err != nil {
		return err
	}

	if strings.TrimSpace(wgInterface) == "" {
		return nil
	}

	link, err := netlink.LinkByName(strings.TrimSpace(wgInterface))
	if err != nil {
		return fmt.Errorf("failed to find the interface %s: %v", wgInterface, err)
	}

	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete the interface %s: %v", wgInterface, err)
	}

	return resetDNS(verbose)
}

func setDeviceIp(ip net.IPNet, deviceName string, _ bool) error {
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
