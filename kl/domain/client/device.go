package client

import (
	"fmt"
	"net"

	"github.com/kloudlite/kl/pkg/functions"
)

type AccountVpnConfig struct {
	WGconf     string `json:"wg"`
	DeviceName string `json:"device"`
}

func CurrentDeviceName() (string, error) {
	file, err := GetDeviceContext()
	if err != nil {
		return "", functions.NewE(err)
	}
	if file.DeviceName == "" {
		return "",
			functions.Error("no selected device. please select one using \"kl account switch\"")
	}
	return file.DeviceName, nil
}

func CurrentDeviceIp() (*string, error) {
	dev, err := CurrentDeviceName()
	if err != nil {
		return nil, functions.NewE(err)
	}

	ipAddr, err := net.ResolveIPAddr("", fmt.Sprintf("%s.device.local", dev))
	if err != nil {
		return nil, functions.NewE(err)
	}

	kk := ipAddr.IP.String()
	return &kk, nil
}

func SelectDevice(deviceName string) error {
	file, err := GetDeviceContext()
	if err != nil {
		return functions.NewE(err)
	}

	file.DeviceName = deviceName

	if file.DeviceName == "" {
		return nil
	}

	err = WriteDeviceContext(&DeviceContext{
		DeviceName: deviceName,
	})
	return functions.NewE(err)
}

// func EnsureAppRunning() error {
// 	p, err := proxy.NewProxy(flags.IsDev())
// 	if err != nil {
// 		return functions.NewE(err)
// 	}
//
// 	count := 0
// 	for {
// 		if p.Status() {
// 			return nil
// 		}
//
// 		if runtime.GOOS != "windows" {
// 			cmd := exec.Command("sudo", "echo", "")
// 			cmd.Stdin = os.Stdin
// 			cmd.Stderr = os.Stderr
// 			cmd.Stdout = os.Stdout
//
// 			err := cmd.Run()
// 			if err != nil {
// 				return functions.NewE(err)
// 			}
//
// 			command := exec.Command("sudo", flags.GetCliPath(), "app", "start")
// 			_ = command.Start()
//
// 		} else {
// 			_, err = functions.WinSudoExec(fmt.Sprintf("%s app start", flags.GetCliPath()), nil)
// 			if err != nil {
// 				functions.PrintError(err)
// 			}
// 		}
//
// 		count++
// 		if count >= 2 {
// 			return fmt.Errorf("failed to start app")
// 		}
//
// 		time.Sleep(2 * time.Second)
// 	}
// }
