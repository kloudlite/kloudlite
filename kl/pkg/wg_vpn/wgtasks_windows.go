package vpn

import (
	"errors"
	"net"

	"github.com/kloudlite/kl/pkg/ui/text"
)

const (
	KlWgInterface = "wgkl"
)

func getCurrentDns() ([]string, error) {

	return []string{}, nil
}

func SetDeviceIp(ip net.IPNet, deviceName string, _ bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}

func StartService(_ string, verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}

func ipRouteAdd(ip string, interfaceIp string, devName string, verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}

func StopService(verbose bool) error {
	return errors.New(
		text.Colored("This command is not available for windows, will be available soon", 209),
	)
}
