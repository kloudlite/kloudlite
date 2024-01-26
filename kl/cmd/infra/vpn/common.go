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

	device, err := server.GetDevice(fn.MakeOption("deviceName", devName))

	if device.Spec.ActiveNamespace == "" {
		fn.Log(fmt.Sprintf("[#] no namespace selected for the device %s, you can select namespace using 'kl infra vpn activate -n <namespace>'\n", devName))
	}

	if len(device.Spec.Ports) == 0 {
		fn.Log(fmt.Sprintf("[#] no ports found for device %s, you can expose ports using 'kl infra vpn expose'\n", devName))
	}

	if device.WireguardConfig.Value == "" {
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
