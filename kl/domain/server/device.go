package server

import (
	"os"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

func GetDevice(options ...fn.Option) (*Device, error) {
	devName := fn.GetOption(options, "deviceName")

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getCoreDevice", map[string]any{
		"name": devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Device](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func GetDeviceName(devName string) (*CheckName, error) {
	_, err := EnsureAccount()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_CoreCheckNameAvailability", map[string]any{
		"resType": VPNDeviceType,
		"name":    devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func CreateDevice(selectedDeviceName string, devName string) (*Device, error) {
	_, err := EnsureAccount()
	if err != nil {
		return nil, err
	}
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_createCoreDevice", map[string]any{
		"vpnDevice": map[string]any{
			"displayName": devName,
			"metadata": map[string]any{
				"name": selectedDeviceName,
			},
		},
	}, &cookie)
	if err != nil {
		return nil, err
	}
	if fromResp, err := GetFromResp[Device](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func UpdateDevice(ports []DevicePort) error {

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	device, err := GetDevice([]fn.Option{
		fn.MakeOption("deviceName", devName),
	}...)

	for _, port := range ports {
		matched := false

		for i, port2 := range device.Spec.Ports {
			if port2.Port == port.Port {
				matched = true
				device.Spec.Ports[i] = port
				break
			}
		}

		if !matched {
			device.Spec.Ports = append(device.Spec.Ports, port)
		}
	}

	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_updateCoreDevicePorts", map[string]any{
		"deviceName": devName,
		"ports":      device.Spec.Ports,
	}, &cookie)

	if err != nil {
		return err
	}

	_, err = GetFromResp[bool](respData)
	if err != nil {
		return err
	}
	return nil
}

func DeleteDevicePort(ports []DevicePort) error {
	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	device, err := GetDevice([]fn.Option{
		fn.MakeOption("deviceName", devName),
	}...)

	newPorts := make([]DevicePort, 0)
	for _, port := range device.Spec.Ports {
		matched := false
		for _, port2 := range ports {
			if port.Port == port2.Port {
				matched = true
				break
			}
		}

		if !matched {
			newPorts = append(newPorts, port)
		}
	}

	device.Spec.Ports = newPorts

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_CoreUpdateDevicePorts", map[string]any{
		"deviceName": devName,
		"ports":      device.Spec.Ports,
	}, &cookie)

	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	}

	return nil
}

func EnsureDevice(options ...fn.Option) (string, error) {
	devName := fn.GetOption(options, "deviceName")

	if devName == "" {
		currDevName, _ := client.CurrentDeviceName()
		if currDevName != "" {
			devName = currDevName
		}
	}

	if devName != "" {
		dev, err := GetDevice(fn.MakeOption("deviceName", devName))

		if err == nil {
			return dev.Metadata.Name, nil
		}
	}

	if devName != "" {
		var err error
		devName, err = os.Hostname()
		if err != nil {
			return "", err
		}
	}

	devResult, err := GetDeviceName(devName)
	if err != nil {
		return "", err
	}

	selectedDeviceName := ""
	if devResult.Result == true {
		selectedDeviceName = devName
	} else {
		deviceName, err := fzf.FindOne(
			devResult.SuggestedNames,
			func(deviceName string) string {
				return deviceName
			},
			fzf.WithPrompt("Select Device Name > "),
		)
		if err != nil {
			return "", err
		}

		selectedDeviceName = *deviceName
	}

	dev, err := CreateDevice(selectedDeviceName, devName)
	if err != nil {
		return "", err
	}

	fn.Logf("Device created: %s", dev.Metadata.Name)
	client.WriteDeviceContext(dev.Metadata.Name)

	return dev.Metadata.Name, nil
}

func UpdateDeviceEnv(options ...fn.Option) error {

	envName := fn.GetOption(options, "envName")
	projectName := fn.GetOption(options, "projectName")

	var err error
	projectName, err = EnsureProject(options...)
	if err != nil {
		return err
	}

	env, err := EnsureEnv(&client.Env{
		Name: envName,
	}, options...)
	if err != nil {
		return err
	}

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}
	respData, err := klFetch("cli_CoreUpdateDeviceEnv", map[string]any{
		"deviceName":  devName,
		"envName":     env.Name,
		"projectName": projectName,
	}, &cookie)
	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	}

	return nil
}
