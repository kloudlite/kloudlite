package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/server"
)

const (
	KL_WG_INTERFACE = "wgkl"
)

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
		color.ColorText("This command is not availabel for windows, will be available soon", 209),
	)
}

func ipRouteAdd(ip string, interfaceIp string, verbose bool) error {
	return nil
}

func stopService(verbose bool) error {
	return errors.New(
		color.ColorText("This command is not availabel for windows, will be available soon", 209),
	)
}
