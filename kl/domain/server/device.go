package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/kloudlite/kl/domain/client"
	"github.com/miekg/dns"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type Device struct {
	AccountName       string `json:"accountName"`
	CreationTime      string `json:"creationTime"`
	CreatedBy         User   `json:"createdBy"`
	DisplayName       string `json:"displayName"`
	GlobalVPNName     string `json:"globalVPNName"`
	ID                string `json:"id"`
	IPAddress         string `json:"ipAddr"`
	LastUpdatedBy     User   `json:"lastUpdatedBy"`
	MarkedForDeletion bool   `json:"markedForDeletion"`
	Metadata          struct {
		Annotations       map[string]string `json:"annotations"`
		CreationTimestamp string            `json:"creationTimestamp"`
		DeletionTimestamp string            `json:"deletionTimestamp"`
		Labels            map[string]string `json:"labels"`
		Name              string            `json:"name"`
	} `json:"metadata"`
	PrivateKey      string `json:"privateKey"`
	PublicEndpoint  string `json:"publicEndpoint"`
	PublicKey       string `json:"publicKey"`
	UpdateTime      string `json:"updateTime"`
	WireguardConfig struct {
		Value    string `json:"value"`
		Encoding string `json:"encoding"`
	} `json:"wireguardConfig,omitempty"`
}

const (
	Default_GVPN = "default"
)

type DeviceList struct {
	Edges Edges[Env] `json:"edges"`
}

func GetVPNDevice(devName string, options ...fn.Option) (*Device, error) {
	cookie, err := getCookie(options...)
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getGlobalVpnDevice", map[string]any{
		"gvpn":       Default_GVPN,
		"deviceName": devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	return GetFromResp[Device](respData)
}

func createDevice(devName string) (*Device, error) {
	cn, err := getDeviceName(devName)
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	dn := devName
	if !cn.Result {
		if len(cn.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for device %s", devName)
		}

		dn = cn.SuggestedNames[0]
	}

	fn.Logf("creating new device %s", dn)
	respData, err := klFetch("cli_createGlobalVPNDevice", map[string]any{
		"gvpnDevice": map[string]any{
			"metadata":       map[string]string{"name": dn},
			"globalVPNName":  Default_GVPN,
			"displayName":    dn,
			"creationMethod": "kl",
		},
	}, &cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to create vpn: %s", err.Error())
	}

	d, err := GetFromResp[Device](respData)
	if err != nil {
		return nil, err
	}

	if err := client.SelectDevice(d.Metadata.Name); err != nil {
		return nil, err
	}

	return d, nil
}

// func EnsureDevice(options ...fn.Option) (*Device, error) {
// 	dc, err := client.GetDeviceContext()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	hostName, err := os.Hostname()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if dc.DeviceName == "" {
// 		return createDevice(hostName)
// 	}
//
// 	d, err := getVPNDevice(dc.DeviceName, options...)
// 	if err != nil {
// 		fn.Warnf("failed to get VPN device: %s", err.Error())
// 		return createDevice(hostName)
// 	}
//
// 	return d, nil
// }

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "global_vpn_device"
)

func CheckDeviceStatus() bool {
	verbose := false

	logF := func(format string, v ...interface{}) {
		if verbose {
			if len(v) > 0 {
				fn.Log(format, v)
			} else {
				fn.Log(format)
			}
		}
	}

	s, err := client.GetDeviceContext()
	if err != nil {
		logF(err.Error())
		return false
	}

	if len(s.DeviceDns) == 0 {
		logF("No DNS record found for device")
		return false
	}

	dnsServer := s.DeviceDns[0]

	client := new(dns.Client)

	client.Timeout = 2 * time.Second

	// Create a new DNS message
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn("one.one.one.one"), dns.TypeA)
	message.RecursionDesired = true

	// Send the DNS query
	response, _, err := client.Exchange(message, dnsServer+":53")
	if err != nil {
		logF("Failed to get DNS response: %v\n", err)
		return false
	}

	// Print the response
	if response.Rcode != dns.RcodeSuccess {
		logF("Query failed: %s\n", dns.RcodeToString[response.Rcode])
		return false
	} else {
		for _, answer := range response.Answer {
			logF("%s\n", answer.String())
		}

	}

	return true
}

func getDeviceName(devName string) (*CheckName, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
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

func createVpnForAccount() (*Device, error) {
	devName, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	checkNames, err := getDeviceName(devName)
	if err != nil {
		return nil, err
	}
	if !checkNames.Result {
		if len(checkNames.SuggestedNames) == 0 {
			return nil, fmt.Errorf("no suggested names for device %s", devName)
		}
		devName = checkNames.SuggestedNames[0]
	}
	device, err := createDevice(devName)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func GetAccVPNConfig(account string) (*client.AccountVpnConfig, error) {
	cfgFolder, err := client.GetConfigFolder()
	if err != nil {
		return nil, fn.NewE(err)
	}
	err = os.MkdirAll(path.Join(cfgFolder, "vpn"), 0755)
	if err != nil {
		return nil, fn.NewE(err)
	}
	cfgPath := path.Join(cfgFolder, "vpn", fmt.Sprintf("%s.json", account))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		dev, err := createVpnForAccount()
		if err != nil {
			return nil, fn.NewE(err)
		}
		accountVpnConfig := client.AccountVpnConfig{
			WGconf:     dev.WireguardConfig.Value,
			DeviceName: dev.Metadata.Name,
		}
		marshal, err := json.Marshal(accountVpnConfig)
		if err != nil {
			return nil, fn.NewE(err)
		}
		err = os.WriteFile(cfgPath, marshal, 0644)
		if err != nil {
			return nil, fn.NewE(err)
		}
	}

	var accVPNConfig client.AccountVpnConfig
	c, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fn.Error("failed to read vpn config")
	}
	err = json.Unmarshal(c, &accVPNConfig)
	if err != nil {
		return nil, fn.Error("failed to parse vpn config")
	}

	if accVPNConfig.WGconf == "" {
		d, err := GetVPNDevice(accVPNConfig.DeviceName, fn.MakeOption("accountName", account))
		if err != nil {
			return nil, fn.NewE(err)
		}

		accVPNConfig.WGconf = d.WireguardConfig.Value

		marshal, err := json.Marshal(accVPNConfig)
		if err != nil {
			return nil, fn.NewE(err)
		}
		err = os.WriteFile(cfgPath, marshal, 0644)
		if err != nil {
			return nil, fn.NewE(err)
		}
	}

	return &accVPNConfig, nil
}
