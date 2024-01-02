package server

import (
	"errors"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Device struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
	Spec        struct {
		CnameRecords []struct {
			Host   string `json:"host"`
			Target string `json:"target"`
		} `json:"cnameRecords"`
		DeviceNamespace string `json:"deviceNamespace"`
		Ports           []struct {
			Port       int `json:"port"`
			TargetPort int `json:"targetPort,omitempty"`
		} `json:"ports"`
	} `json:"spec"`
	WireguardConfig *struct {
		Encoding string `json:"encoding"`
		Value    string `json:"value"`
	} `json:"wireguardConfig,omitempty"`
}

func ListDevices(options ...fn.Option) ([]Device, error) {
	clusterName := fn.GetOption(options, "clusterName")
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	if clusterName != "" {

		c, err := SelectCluster(clusterName)
		if err != nil {
			return nil, err
		}
		clusterName = c.Metadata.Name
	}

	s, err := client.CurrentClusterName()
	if err != nil {
		return nil, err
	}

	clusterName = s

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
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	s, err := client.CurrentClusterName()
	if err != nil {
		return nil, err
	}
	clusterName := s

	if devName != "" {
		if s, err := client.CurrentDeviceName(); err != nil {
			return nil, err
		} else {
			devName = s
		}
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

func SelectDevice(deviceName string) (*Device, error) {
	persistSelectedDevice := func(deviceName string) error {
		err := client.SelectDevice(deviceName)
		if err != nil {
			return err
		}
		return nil
	}

	if deviceName != "" {
		if s, err := client.CurrentDeviceName(); err == nil {
			deviceName = s
		}
	}

	devices, err := ListDevices()
	if err != nil {
		return nil, err
	}

	if deviceName != "" {
		for _, d := range devices {
			if d.Metadata.Name == deviceName {
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

// type DAccount struct {
// 	Devices []Device `json:"devices"`
// }
//
// type DApp struct {
// 	ReadableId string `json:"readableId"`
// 	Name       string `json:"name"`
// 	Id         string `json:"id"`
// }
//
// type Port struct {
// 	Port       int `json:"port"`
// 	TargetPort int `json:"targetPort,omitempty"`
// }
//
// type Device struct {
// 	Region        string            `json:"region"`
// 	Ports         []Port            `json:"ports"`
// 	Configuration map[string]string `json:"configuration"`
// 	Name          string            `json:"name"`
// 	Id            string            `json:"id"`
// 	Intercepted   []DApp            `json:"interceptingServices"`
// }
//
// func GetDevice(deviceId string) (*Device, error) {
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	respData, err := klFetch("cli_getDevice", map[string]any{
// 		"deviceId": deviceId,
// 	}, &cookie)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	type Response struct {
// 		Device Device `json:"data"`
// 	}
//
// 	// fmt.Println(string(respData),"here")
// 	var resp Response
// 	err = json.Unmarshal(respData, &resp)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &resp.Device, nil
// }
//
// func GetDevices(options ...fn.Option) ([]Device, error) {
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	accountId := fn.GetOption(options, "accountId")
// 	if accountId == "" {
// 		accountId, err = client.CurrentAccountName()
//
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
//
// 	respData, err := klFetch("cli_getDevices", map[string]any{}, &cookie)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	type Response struct {
// 		Account map[string]DAccount `json:"data"`
// 	}
//
// 	var resp Response
// 	err = json.Unmarshal(respData, &resp)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	account := resp.Account[accountId]
//
// 	return account.Devices, nil
// }
//
// func InterceptApp(devieId, appId string) error {
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return err
// 	}
//
// 	respData, err := klFetch("cli_interceptApp", map[string]any{
// 		"deviceId": devieId,
// 		"appId":    appId,
// 	}, &cookie)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	var resp struct {
// 		Inercepted bool `json:"data"`
// 	}
//
// 	err = json.Unmarshal(respData, &resp)
// 	if err != nil {
// 		return err
// 	}
//
// 	if resp.Inercepted {
// 		return nil
// 	}
//
// 	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
// }
//
// func CloseInterceptApp(appId string) error {
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return err
// 	}
//
// 	respData, err := klFetch("cli_closeIntercept", map[string]any{
// 		"appId": appId,
// 	}, &cookie)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	var resp struct {
// 		Inercepted bool `json:"data"`
// 	}
//
// 	err = json.Unmarshal(respData, &resp)
// 	if err != nil {
// 		return err
// 	}
//
// 	if resp.Inercepted {
// 		return nil
// 	}
//
// 	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
// }
//
// func CreateDevice(deviceName string) error {
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return err
// 	}
//
// 	accId, err := client.CurrentAccountName()
// 	if err != nil {
// 		return err
// 	}
//
// 	_, err = klFetch("cli_createDevice", map[string]any{
// 		"name":      deviceName,
// 		"accountId": accId,
// 	}, &cookie)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func UpdateDevice(ports []Port, region *string) error {
//
// 	if !(region != nil || len(ports) >= 1) {
// 		return errors.New("nothing to change")
// 	}
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return err
// 	}
//
// 	deviceId, err := CurrentDeviceId()
// 	if err != nil {
// 		return err
// 	}
//
// 	devices, err := GetDevices()
// 	if err != nil {
// 		return err
// 	}
//
// 	var activeDevice *Device
//
// 	for i, d := range devices {
// 		if d.Id == deviceId {
// 			dv := devices[i]
// 			activeDevice = &dv
// 		}
// 	}
//
// 	if activeDevice == nil {
// 		return errors.New("selected device is not present in the selected account")
// 	}
//
// 	if region != nil {
// 		activeDevice.Region = *region
// 	}
//
// 	if len(ports) >= 1 {
// 		for _, p := range ports {
// 			matched := false
// 			for i, p2 := range activeDevice.Ports {
// 				if p2.Port == p.Port {
// 					matched = true
// 					activeDevice.Ports[i] = p
// 					break
// 				}
// 			}
//
// 			if !matched {
// 				activeDevice.Ports = append(activeDevice.Ports, p)
// 			}
// 		}
// 	}
//
// 	if region != nil || len(ports) >= 1 {
// 		if _, err = klFetch("cli_updateDevice", map[string]any{
// 			"deviceId": activeDevice.Id,
// 			"name":     activeDevice.Name,
// 			"region":   activeDevice.Region,
// 			"ports":    activeDevice.Ports,
// 		}, &cookie); err != nil {
// 			return err
// 		}
// 	}
//
// 	return nil
// }
//
// func DeleteDevicePort(ports []Port) error {
//
// 	if len(ports) == 0 {
// 		return errors.New("nothing to change")
// 	}
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return err
// 	}
//
// 	deviceId, err := CurrentDeviceId()
// 	if err != nil {
// 		return err
// 	}
//
// 	devices, err := GetDevices()
// 	if err != nil {
// 		return err
// 	}
//
// 	var activeDevice *Device
//
// 	for i, d := range devices {
// 		if d.Id == deviceId {
// 			dv := devices[i]
// 			activeDevice = &dv
// 		}
// 	}
//
// 	if activeDevice == nil {
// 		return errors.New("selected device is not present in the selected account")
// 	}
//
// 	newPorts := make([]Port, 0)
//
// 	for _, p := range activeDevice.Ports {
// 		matched := false
// 		for _, p2 := range ports {
// 			if p.Port == p2.Port {
// 				matched = true
// 				break
// 			}
// 		}
//
// 		if !matched {
// 			newPorts = append(newPorts, p)
// 		}
// 	}
//
// 	if _, err = klFetch("cli_updateDevice", map[string]any{
// 		"deviceId": activeDevice.Id,
// 		"name":     activeDevice.Name,
// 		"region":   activeDevice.Region,
// 		"ports":    newPorts,
// 	}, &cookie); err != nil {
// 		return err
// 	}
//
// 	return nil
// }
