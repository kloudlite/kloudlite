package server

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/pkg/ui/input"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

func ListDevices() ([]Device, error) {

	_, err := EnsureAccount()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listCoreDevices", map[string]any{
		"pq": PaginationDefault,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Device](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func GetDevice(options ...fn.Option) (*Device, error) {
	devName := fn.GetOption(options, "deviceName")

	var err error
	devName, err = EnsureDevice(options...)

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

func SelectDevice(devName string) (*Device, error) {
	persistSelectedDevice := func(deviceName string) error {
		err := client.SelectDevice(deviceName)
		if err != nil {
			return err
		}
		return nil
	}

	devices, err := ListDevices()
	if err != nil {
		return nil, err
	}

	if devName != "" {
		for _, d := range devices {
			if d.Metadata.Name == devName {
				if err := persistSelectedDevice(d.Metadata.Name); err != nil {
					return nil, err
				}
				return &d, nil
			}
		}
		return nil, errors.New("you don't have access to this device")
	}
	if len(devices) == 0 {
		deviceName, err := input.Prompt(input.Options{
			Placeholder: "Enter device name",
			CharLimit:   25,
			Password:    false,
		})
		if err != nil {
			return nil, err
		}
		//suggestedNames, err := GetDeviceName(deviceName)
		//if err != nil {
		//	return nil, err
		//}
		//selectedDeviceName, err := SelectDeviceName(suggestedNames.SuggestedNames)
		//device, err := CreateDevice(selectedDeviceName, deviceName)
		device, err := CreateDevice(deviceName, deviceName)
		if err != nil {
			fn.PrintError(err)
			return nil, err
		}
		fmt.Println(deviceName, "has been created")
		return device, nil
	}

	device, err := fzf.FindOne(
		devices,
		func(device Device) string {
			return fmt.Sprintf("%s (%s)", device.DisplayName, device.Metadata.Name)
		},
		fzf.WithPrompt("Select Device > "),
	)

	if err != nil {
		return nil, err
	}

	if err := persistSelectedDevice(device.Metadata.Name); err != nil {
		return nil, err
	}

	return device, nil
}

func GetDeviceName(devName string) (*CheckName, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_CoreCheckNameAvailability", map[string]any{
		"resType": VPNDeviceType,
		"name":    devName,
	}, &cookie)
	if err != nil {
		fmt.Println(respData, err)
		return nil, err
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func SelectDeviceName(suggestedNames []string) (string, error) {
	deviceName, err := fzf.FindOne(
		suggestedNames,
		func(deviceName string) string {
			return deviceName
		},
		fzf.WithPrompt("Select Device Name > "),
	)

	if err != nil {
		return "", err
	}

	return *deviceName, nil
}

func CreateDevice(selectedDeviceName string, devName string) (*Device, error) {
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

	devName, err := EnsureDevice()
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
	devName, err := EnsureDevice()
	if err != nil {
		return err
	}

	clusterName, err := client.CurrentClusterName()
	if err != nil {
		return err
	}

	device, err := GetDevice([]fn.Option{
		fn.MakeOption("deviceName", devName),
		fn.MakeOption("clusterName", clusterName),
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

	respData, err := klFetch("cli_updateDevicePort", map[string]any{
		"clusterName": clusterName,
		"deviceName":  devName,
		"ports":       device.Spec.Ports,
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

	if devName != "" {
		return devName, nil
	}

	devName, _ = client.CurrentDeviceName()

	if devName != "" {
		return devName, nil
	}

	dev, err := SelectDevice("")

	if err != nil {
		return "", err
	}

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

	devName, err := EnsureDevice(options...)
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_updateDeviceNs", map[string]any{
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
