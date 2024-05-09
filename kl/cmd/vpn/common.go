package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	wg_svc "github.com/kloudlite/kl/pkg/wg_vpn/wg_service"
)

const (
	ifName string = "utun2464"
)

func startConfiguration(verbose bool, options ...fn.Option) error {
	device, err := server.EnsureDevice(options...)
	if err != nil {
		return err
	}

	if device.WireguardConfig.Value == "" {
		return errors.New("no wireguard config found, please try again in few seconds")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
	if err != nil {
		return err
	}

	if runtime.GOOS == constants.RuntimeWindows {
		if err := wg_svc.StartVpn(configuration); err != nil {
			return err
		}

		return nil
	}

	if err := wg_vpn.Configure(configuration, ifName, verbose); err != nil {
		return err
	}

	if wg_vpn.IsSystemdReslov() {
		if err := wg_vpn.ExecCmd(fmt.Sprintf("resolvectl domain %s %s", device.Metadata.Name, func() string {
			if device.EnvironmentName != "" {
				return fmt.Sprintf("%s.svc.cluster.local", device.EnvironmentName)
			}

			return "~."
		}()), false); err != nil {
			return err
		}
	}
	return nil
}
