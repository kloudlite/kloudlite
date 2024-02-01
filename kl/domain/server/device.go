package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type DevicePort struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort,omitempty"`
}

type Device struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
	EnvName     string   `json:"environmentName"`
	ProjectName string   `json:"projectName"`
	ClusterName string   `json:"clusterName"`
	Spec        struct {
		CnameRecords []struct {
			Host   string `json:"host"`
			Target string `json:"target"`
		} `json:"cnameRecords"`
		ActiveNamespace string       `json:"activeNamespace"`
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

func GetDevice(options ...fn.Option) (*Device, error) {
	devName := fn.GetOption(options, "deviceName")

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getDevice", map[string]any{
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

	respData, err := klFetch("cli_coreCheckNameAvailability", map[string]any{
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
	respData, err := klFetch("cli_createDevice", map[string]any{
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

	respData, err := klFetch("cli_updateDevicePorts", map[string]any{
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

	respData, err := klFetch("cli_updateDevicePorts", map[string]any{
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

		if err != nil {
			devName = ""
		}
	}

	var err error
	devName, err = os.Hostname()
	if err != nil {
		return "", err
	}

	devResult, err := GetDeviceName(devName)
	if err != nil {
		return "", err
	}

	selectedDeviceName := ""
	if devResult.Result == true {
		selectedDeviceName = devName
	} else {
		if len(devResult.SuggestedNames) == 0 {
			return "", fmt.Errorf("no suggested names found")
		}

		selectedDeviceName = devResult.SuggestedNames[0]
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

	respData, err := klFetch("cli_updateDeviceEnv", map[string]any{
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

func UpdateDeviceNS(namespace string) error {
	devName, err := EnsureDevice()
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_updateDeviceNs", map[string]any{
		"deviceName": devName,
		"ns":         namespace,
	}, &cookie)

	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	}

	return nil
}

func UpdateDeviceClusterName(clusterName string) error {

	devName, err := client.CurrentDeviceName()
	if err != nil || devName == "" {
		devName, err = EnsureDevice()
		if err != nil {
			return err
		}
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_updateDeviceCluster", map[string]any{
		"clusterName": clusterName,
		"deviceName":  devName,
	}, &cookie)

	if err != nil {
		return err
	}

	if _, err := GetFromResp[bool](respData); err != nil {
		return err
	}

	return nil
}

func CheckDeviceStatus() bool {

	httpClient := http.Client{Timeout: 200 * time.Millisecond}
	if _, err := httpClient.Get("http://10.13.0.1:17171"); err != nil {
		return false
	}

	return true
}
