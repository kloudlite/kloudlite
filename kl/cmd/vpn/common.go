package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	wg_svc "github.com/kloudlite/kl/pkg/wg_vpn/wg_service"
)

const (
	ifName string = "utun2464"
)

func startConfiguration(verbose bool, options ...fn.Option) error {
	selectedDevice, err := client.GetDeviceContext()
	if err != nil {
		return err
	}

	devName := selectedDevice.DeviceName

	// ensure device
	if err := func() error {
		if devName != "" {
			return nil
		}

		// fetch device
		devs, err := server.ListVPNDevice(options...)
		if err != nil {
			return err
		}

		if len(devs) == 0 {
			return fmt.Errorf("no device found")
		}

		if len(devs) == 1 {
			devName = devs[0].Metadata.Name
			return nil
		}

		devName = devs[0].Metadata.Name

		dev, err := fzf.FindOne(
			devs,
			func(item server.Device) string { return item.Metadata.Name },
			fzf.WithPrompt("Select VPN device >"),
		)

		if err != nil {
			return err
		}

		devName = dev.Metadata.Name
		return nil
	}(); err != nil {
		return err
	}

	if selectedDevice.DeviceName != devName {
		if err := client.SelectDevice(devName); err != nil {
			return err
		}
	}

	// if !skipCheck {
	// 	switch flags.CliName {
	// 	case constants.CoreCliName:
	// 		envName := fn.GetOption(options, "envName")
	// 		if envName != "" {
	// 			en, err := client.CurrentEnv()
	//
	// 			if (err == nil && en.Name != envName) || (err != nil && envName != "") {
	// 				_, err := server.GetVPNDevice(devName, options...)
	// 				if err != nil {
	// 					return err
	// 				}
	// 			}
	// 		}
	//
	// 	case constants.InfraCliName:
	// 		_, err := server.GetVPNDevice(devName, options...)
	// 		if err != nil {
	// 			return err
	// 		}
	// 		//clusterName := fn.GetOption(options, "clusterName")
	// 		//if clusterName != "" {
	// 		//	cn, err := client.CurrentClusterName()
	// 		//	if err != nil {
	// 		//		return err
	// 		//	}
	// 		//	if cn != "" && cn != clusterName {
	// 		//		if err := server.UpdateDeviceClusterName(clusterName); err != nil {
	// 		//			return err
	// 		//		}
	// 		//	}
	// 		//
	// 		//	time.Sleep(2 * time.Second)
	// 		//}
	// 	}
	// }

	device, err := server.GetVPNDevice(devName, options...)
	if err != nil {
		switch flags.CliName {
		case constants.CoreCliName:
			return err
		//case constants.InfraCliName:
		//	return err
		default:
			return err
		}
	}

	//if device.ClusterName != "" {
	//	_ = client.SetDevInfo(fn.Truncate(device.ClusterName, 15))
	//} else {
	//	_ = client.SetDevInfo(fmt.Sprintf("%s/%s", fn.Truncate(device.ProjectName, 5), fn.Truncate(device.EnvName, 5)))
	//}

	//if !skipCheck {
	//	switch flags.CliName {
	//	case constants.CoreCliName:
	//		envName := fn.GetOption(options, "envName")
	//		//projectName := fn.GetOption(options, "projectName")
	//
	//		if envName == "" {
	//			en, err := client.CurrentEnv()
	//			if err == nil && en.Name != "" {
	//				envName = en.Name
	//			}
	//		}
	//		if (envName != "" && device.Metadata. != envName) {
	//			if err := server.UpdateDeviceEnv([]fn.Option{
	//				fn.MakeOption("envName", envName),
	//				fn.MakeOption("projectName", projectName),
	//			}...); err != nil {
	//				return err
	//			}
	//			time.Sleep(2 * time.Second)
	//		}

	//case constants.InfraCliName:
	//	clusterName := fn.GetOption(options, "clusterName")
	//
	//	if clusterName == "" {
	//		if s, err := client.CurrentClusterName(); err != nil {
	//			return err
	//		} else {
	//			clusterName = s
	//		}
	//	}
	//
	//	if device.ClusterName == "" || (device.ClusterName != clusterName) {
	//		if err := server.UpdateDeviceClusterName(clusterName); err != nil {
	//			return err
	//		}
	//
	//		time.Sleep(2 * time.Second)
	//	}
	//}
	//}

	//if len(device.Spec.Ports) == 0 {
	//	fn.Log(text.Yellow(fmt.Sprintf("[#] no ports found for device %s, you can export ports using %s vpn expose\n", devName, flags.CliName)))
	//}

	// fmt.Println(device)
	// fmt.Println(device.WireguardConfig)

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

	if err := wg_vpn.Configure(configuration, devName, ifName, verbose); err != nil {
		return err
	}

	if wg_vpn.IsSystemdReslov() {
		if err := wg_vpn.ExecCmd(fmt.Sprintf("resolvectl domain %s %s", device.Metadata.Name, func() string {
			if device.Metadata.Namespace != "" {
				return fmt.Sprintf("%s.svc.cluster.local", device.Metadata.Namespace)
			}

			return "~."
		}()), false); err != nil {
			return err
		}
	}
	return nil
}
