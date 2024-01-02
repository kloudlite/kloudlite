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

func setDeviceIp(deviceIp string, deviceName string, verbose bool) error {
	return execCmd(fmt.Sprintf("ifconfig %s %s %s", deviceName, deviceIp, deviceIp), verbose)
}

func startService(verbose bool) error {

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	if err := execCmd(fmt.Sprintf("ip link add dev %s type wireguard", devName), verbose); err != nil {
		return err
	}

	return execCmd(fmt.Sprintf("ip link set mtu 1420 up dev %s", devName), verbose)

}

func ipRouteAdd(ip string, interfaceIp string, devName string, verbose bool) error {
	return execCmd(fmt.Sprintf("ip -4 route add %s dev %s", ip, devName), verbose)
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

	if err := execCmd(fmt.Sprintf("ip link del dev %s", strings.TrimSpace(wgInterface)), verbose); err != nil {
		return err
	}

	return resetDNS(verbose)
}
