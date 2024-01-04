package wg

import (
	"errors"
	"github.com/kloudlite/kl/pkg/ui/text"
	"net"
)

const (
	KlWgInterface = "wgkl"
)

func configureDarwin(_ string, _ bool) error {
	// not required to implement
	return nil
}

func connect(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}
func disconnect(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}

func setDNS(dns []net.IP, verbose bool) error {
	return nil
}
func resetDNS(verbose bool) error {
	return nil
}

func setDeviceIp(deviceIp string, verbose bool) error {
	return nil
}

func startService(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}

func ipRouteAdd(ip string, interfaceIp string, verbose bool) error {
	return nil
}

func stopService(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}
