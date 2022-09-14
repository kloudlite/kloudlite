package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
)

const (
	KL_WG_INTERFACE = "wgkl"
)

func setDNS(dns []net.IP, verbose bool) error {
	if verbose {
		common.Log("[#] setting /etc/resolv.conf")
	}
	file, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return err
	}

	if _, e := os.Stat("/etc/resolv.conf.back"); errors.Is(e, os.ErrNotExist) {
		e = os.WriteFile("/etc/resolv.conf.back", file, 0644)
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

		server.SetActiveDns(dnsArr)

		return resolveString
	}()), 0644)

	return err
}
func resetDNS(verbose bool) error {
	if verbose {
		common.Log("[#] resetting /etc/resolv.conf")
	}

	err := os.Remove("/etc/resolv.conf")
	if err != nil {
		return err
	}
	err = os.Rename("/etc/resolv.conf.back", "/etc/resolv.conf")
	return err
}

func setDeviceIp(deviceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("ifconfig %s %s %s", KL_WG_INTERFACE, deviceIp, deviceIp), verbose)
}

func startService(verbose bool) error {
	err := execCmd(fmt.Sprintf("ip link add dev %s type wireguard", KL_WG_INTERFACE), verbose)
	if err != nil {
		return err
	}

	return execCmd(fmt.Sprintf("ip link set mtu 1420 up dev %s", KL_WG_INTERFACE), verbose)
}

func ipRouteAdd(ip string, interfaceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("ip -4 route add %s dev %s", ip, KL_WG_INTERFACE), verbose)
}

func stopService(verbose bool) error {
	err := execCmd(fmt.Sprintf("ip link del dev %s", KL_WG_INTERFACE), verbose)
	if err != nil {
		return err
	}

	return resetDNS(verbose)
}
