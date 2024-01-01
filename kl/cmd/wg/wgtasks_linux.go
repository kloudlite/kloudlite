package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/util"
)

const (
	KlWgInterface = "wgkl"
)

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
		common_util.Log("[#] setting /etc/resolv.conf")
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

		util.SetActiveDns(dnsArr)

		return resolveString
	}()), 0644)

	return err
}
func resetDNS(verbose bool) error {
	if verbose {
		common_util.Log("[#] resetting /etc/resolv.conf")
	}

	// err := os.Remove("/etc/resolv.conf")
	// if err != nil {
	// 	return err
	// }
	err := os.Rename("/etc/resolv.conf.back", "/etc/resolv.conf")
	return err
}

func setDeviceIp(deviceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("ifconfig %s %s %s", KlWgInterface, deviceIp, deviceIp), verbose)
}

func startService(verbose bool) error {
	err := execCmd(fmt.Sprintf("ip link add dev %s type wireguard", KlWgInterface), verbose)
	if err != nil {
		return err
	}

	return execCmd(fmt.Sprintf("ip link set mtu 1420 up dev %s", KlWgInterface), verbose)
}

func ipRouteAdd(ip string, interfaceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("ip -4 route add %s dev %s", ip, KlWgInterface), verbose)
}

func stopService(verbose bool) error {
	err := execCmd(fmt.Sprintf("ip link del dev %s", KlWgInterface), verbose)
	if err != nil {
		return err
	}

	return resetDNS(verbose)
}
