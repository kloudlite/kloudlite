package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

const (
	ifName string = "utun2464"
)

func startConfiguration(verbose bool) error {

	devName, err := client.CurrentInfraDeviceName()
	if err != nil {
		return err
	}

	device, err := server.GetInfraDevice(fn.MakeOption("deviceName", devName))

	if device.Spec.ActiveNamespace == "" {
		return errors.New(fmt.Sprintf("no env name found for device %s, please use env using kl env switch\n", devName))
	}
	if len(device.Spec.Ports) == 0 {
		return errors.New(fmt.Sprintf("no ports found for device %s, please export ports using kl vpn expose\n", devName))
	}
	if device.WireguardConfig == nil {
		return errors.New("no wireguard config found")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
	if err != nil {
		return err
	}

	return wg_vpn.Configure(configuration, devName, func() string {
		if runtime.GOOS == "darwin" {
			return ifName
		}
		return devName
	}(), verbose)
}
