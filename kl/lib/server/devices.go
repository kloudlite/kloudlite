package server

import (
	"encoding/json"
	"errors"

	"github.com/kloudlite/kl/lib/common"
)

type DAccount struct {
	Devices []Device `json:"devices"`
}

type DApp struct {
	ReadableId string `json:"readableId"`
	Name       string `json:"name"`
	Id         string `json:"id"`
}

type Port struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort,omitempty"`
}

type Device struct {
	Region        string            `json:"region"`
	Ports         []Port            `json:"ports"`
	Configuration map[string]string `json:"configuration"`
	Name          string            `json:"name"`
	Id            string            `json:"id"`
	Intercepted   []DApp            `json:"interceptingServices"`
}

func GetDevice(deviceId string) (*Device, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getDevice", map[string]any{
		"deviceId": deviceId,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Device Device `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Device, nil
}

func GetDevices(options ...common.Option) ([]Device, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	accountId := common.GetOption(options, "accountId")
	if accountId == "" {
		accountId, err = CurrentAccountId()

		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_getDevices", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Account map[string]DAccount `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	account := resp.Account[accountId]

	return account.Devices, nil
}

func InterceptApp(devieId, appId string) error {
	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_interceptApp", map[string]any{
		"deviceId": devieId,
		"appId":    appId,
	}, &cookie)

	if err != nil {
		return err
	}

	var resp struct {
		Inercepted bool `json:"data"`
	}

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return err
	}

	if resp.Inercepted {
		return nil
	}

	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
}

func CloseInterceptApp(appId string) error {
	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_closeIntercept", map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return err
	}

	var resp struct {
		Inercepted bool `json:"data"`
	}

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return err
	}

	if resp.Inercepted {
		return nil
	}

	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
}

func CreateDevice(deviceName string) error {

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	accId, err := CurrentAccountId()
	if err != nil {
		return err
	}

	_, err = klFetch("cli_createDevice", map[string]any{
		"name":      deviceName,
		"accountId": accId,
	}, &cookie)

	if err != nil {
		return err
	}

	return nil
}

func UpdateDevice(ports []Port, region *string) error {

	if !(region != nil || len(ports) >= 1) {
		return errors.New("nothing to change")
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	deviceId, err := CurrentDeviceId()
	if err != nil {
		return err
	}

	devices, err := GetDevices()
	if err != nil {
		return err
	}

	var activeDevice *Device

	for i, d := range devices {
		if d.Id == deviceId {
			dv := devices[i]
			activeDevice = &dv
		}
	}

	if activeDevice == nil {
		return errors.New("selected device is not present in the selected account")
	}

	if region != nil {
		activeDevice.Region = *region
	}

	if len(ports) >= 1 {
		for _, p := range ports {
			matched := false
			for i, p2 := range activeDevice.Ports {
				if p2.Port == p.Port {
					matched = true
					activeDevice.Ports[i] = p
					break
				}
			}

			if !matched {
				activeDevice.Ports = append(activeDevice.Ports, p)
			}
		}
	}

	if region != nil || len(ports) >= 1 {
		if _, err = klFetch("cli_updateDevice", map[string]any{
			"deviceId": activeDevice.Id,
			"name":     activeDevice.Name,
			"region":   activeDevice.Region,
			"ports":    activeDevice.Ports,
		}, &cookie); err != nil {
			return err
		}
	}

	return nil
}

func DeleteDevicePort(ports []Port) error {

	if len(ports) == 0 {
		return errors.New("nothing to change")
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	deviceId, err := CurrentDeviceId()
	if err != nil {
		return err
	}

	devices, err := GetDevices()
	if err != nil {
		return err
	}

	var activeDevice *Device

	for _, d := range devices {
		if d.Id == deviceId {
			activeDevice = &d
		}
	}

	if activeDevice == nil {
		return errors.New("selected device is not present in the selected account")
	}

	newPorts := make([]Port, 0)

	for _, p := range activeDevice.Ports {
		matched := false
		for _, p2 := range ports {
			if p.Port == p2.Port {
				matched = true
				break
			}
		}

		if !matched {
			newPorts = append(newPorts, p)
		}
	}

	if _, err = klFetch("cli_updateDevice", map[string]any{
		"deviceId": activeDevice.Id,
		"name":     activeDevice.Name,
		"region":   activeDevice.Region,
		"ports":    newPorts,
	}, &cookie); err != nil {
		return err
	}

	return nil
}
