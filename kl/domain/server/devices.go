package server

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type DevicePort struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort,omitempty"`
}

type Device struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
	Spec        struct {
		CnameRecords []struct {
			Host   string `json:"host"`
			Target string `json:"target"`
		} `json:"cnameRecords"`
		DeviceNamespace string       `json:"deviceNamespace"`
		Ports           []DevicePort `json:"ports"`
	} `json:"spec"`
	WireguardConfig *struct {
		Encoding string `json:"encoding"`
		Value    string `json:"value"`
	} `json:"wireguardConfig,omitempty"`
}

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "vpn_device"
)

func ListDevices(options ...fn.Option) ([]Device, error) {

	clusterName := fn.GetOption(options, "clusterName")

	var err error
	if clusterName, err = EnsureCluster(options...); err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listDevices", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"clusterName": clusterName,
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

	clusterName, err := EnsureCluster(options...)
	if err != nil {
		return nil, err
	}

	devName, err = EnsureDevice(options...)

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getDevice", map[string]any{
		"clusterName": clusterName,
		"name":        devName,
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

	device, err := fzf.FindOne(
		devices,
		func(device Device) string {
			return device.DisplayName
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
	clusterName, err := EnsureCluster()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType":     VPNDeviceType,
		"clusterName": clusterName,
		"name":        devName,
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
	clusterName, err := EnsureCluster()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_createDevice", map[string]any{
		"clusterName": clusterName,
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

	clusterName, err := client.CurrentClusterName()
	if err != nil {
		return err
	}

	device, err := GetDevice([]fn.Option{
		fn.MakeOption("deviceName", devName),
		fn.MakeOption("clusterName", clusterName),
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

	respData, err := klFetch("cli_updateDevicePort", map[string]any{
		"clusterName": clusterName,
		"deviceName":  devName,
		"ports":       device.Spec.Ports,
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

	_, err := EnsureCluster(options...)
	if err != nil {
		return "", err
	}

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

func UpdateDeviceNS(namespace string) error {
	devName, err := EnsureDevice()
	if err != nil {
		return err
	}

	clusterName, err := client.CurrentClusterName()
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_updateDeviceNs", map[string]any{
		"clusterName": clusterName,
		"deviceName":  devName,
		"namespace":   namespace,
	}, &cookie)

	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	}

	return nil
}
