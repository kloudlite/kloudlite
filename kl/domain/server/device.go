package server

import (
	"fmt"
	"os"
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
	// TODO: match with api (envname)
	EnvironmentName string `json:"environmentName"`
	Metadata        struct {
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

	respData, err := klFetch("cli_createGlobalVPNDevice", map[string]any{
		"gvpnDevice": map[string]any{
			"metadata":       map[string]string{"name": dn},
			"globalVPNName":  Default_GVPN,
			"displayName":    dn,
			"creationMethod": "kl",
		},
	}, &cookie)
	if err != nil {
		return nil, err
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

func EnsureDevice(options ...fn.Option) (*Device, error) {
	dc, err := client.GetDeviceContext()
	if err != nil {
		return nil, err
	}

	hostName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	if dc.DeviceName == "" {
		return createDevice(hostName)
	}

	d, err := getVPNDevice(dc.DeviceName, options...)
	if err != nil {
		fn.Warnf("Failed to get VPN device: %s", err)
		return createDevice(hostName)
	}

	return d, nil
}

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "vpn_device"
)

func getDeviceName(devName string) (*CheckName, error) {
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

func getVPNDevice(devName string, options ...fn.Option) (*Device, error) {
	accountName := fn.GetOption(options, "accountName")
	envName := fn.GetOption(options, "envName")

	accountName, err := EnsureAccount(options...)
	if err != nil {
		return nil, err
	}

	if envName == "" {
		env, err := EnsureEnv(nil, options...)
		if err != nil {
			return nil, err
		}
		envName = env.Name
	}

	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getGlobalVpnDevice", map[string]any{
		// TODO: add env name when api is available
		//"envName":    envName,
		"gvpn":       Default_GVPN,
		"deviceName": devName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	return GetFromResp[Device](respData)
}

func CheckDeviceStatus() bool {
	verbose := false

	logF := func(format string, v ...interface{}) {
		if verbose {
			fn.Logf(format, v)
		}
	}

	s, err := client.GetDeviceDns()
	if err != nil {
		logF(err.Error())
		return false
	}

	if s == "" {
		logF("No DNS record found for device")
		return false
	}

	dnsServer := s

	client := new(dns.Client)

	client.Timeout = 2 * time.Second

	// Create a new DNS message
	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn("kube-dns.kube-system.svc.cluster.local"), dns.TypeA)
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
