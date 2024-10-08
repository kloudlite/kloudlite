package apiclient

import (
	"os"

	"github.com/kloudlite/kl/constants"

	"time"

	"github.com/kloudlite/kl/domain/envclient"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/miekg/dns"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type Device struct {
	TeamName          string `json:"accountName"`
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

func (apic *apiClient) GetVPNDevice(teamName string, devName string) (*Device, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getGlobalVpnDevice", map[string]any{
		"gvpn":       Default_GVPN,
		"deviceName": devName,
	}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	return GetFromResp[Device](respData)
}

func (apic *apiClient) CreateDevice(devName, displayName, team string) (*Device, error) {
	cn, err := getDeviceName(devName, team)
	if err != nil {
		return nil, fn.NewE(err)
	}

	cookie, err := getCookie(fn.MakeOption("teamName", team))
	if err != nil {
		return nil, fn.NewE(err)
	}

	dn := devName
	if !cn.Result {
		if len(cn.SuggestedNames) == 0 {
			return nil, fn.Errorf("no suggested names for device %s", devName)
		}

		dn = cn.SuggestedNames[0]
	}
	fn.Logf("creating new device %s\n", dn)
	respData, err := klFetch("cli_createGlobalVPNDevice", map[string]any{
		"gvpnDevice": map[string]any{
			"metadata":       map[string]string{"name": dn},
			"globalVPNName":  Default_GVPN,
			"displayName":    displayName,
			"creationMethod": "kl",
		},
	}, &cookie)
	if err != nil {
		return nil, fn.Errorf("failed to create vpn: %s", err.Error())
	}

	d, err := GetFromResp[Device](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	return d, nil
}

type CheckName struct {
	Result         bool     `json:"result"`
	SuggestedNames []string `json:"suggestedNames"`
}

const (
	VPNDeviceType = "global_vpn_device"
	ClusterType   = "byok_cluster"
)

func (apic *apiClient) CheckDeviceStatus() bool {
	if !envclient.InsideBox() {
		return false
	}

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

	client := new(dns.Client)
	client.Timeout = 2 * time.Second

	message := new(dns.Msg)
	message.SetQuestion(dns.Fqdn("team.kloudlite.local"), dns.TypeA)
	message.RecursionDesired = true

	// Send the DNS query
	response, _, err := client.Exchange(message, constants.KLDNS+":53")
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

func getDeviceName(devName, team string) (*CheckName, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", team))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType": VPNDeviceType,
		"name":    devName,
	}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}

func (apic *apiClient) CreateVpnForTeam(team string) (*Device, error) {
	devName, err := os.Hostname()
	if err != nil {
		return nil, fn.NewE(err)
	}

	checkNames, err := getDeviceName(devName, team)
	if err != nil {
		return nil, fn.NewE(err)
	}
	if !checkNames.Result {
		if len(checkNames.SuggestedNames) == 0 {
			return nil, fn.Errorf("no suggested names for device %s", devName)
		}
		devName = checkNames.SuggestedNames[0]
	}
	device, err := apic.CreateDevice(devName, devName, team)
	if err != nil {
		return nil, fn.NewE(err)
	}
	return device, nil
}

func (apic *apiClient) GetAccVPNConfig(team string) (*fileclient.TeamVpnConfig, error) {

	avc, err := apic.fc.GetVpnTeamConfig(team)

	if err != nil && os.IsNotExist(err) {
		dev, err := apic.CreateVpnForTeam(team)
		if err != nil {
			return nil, fn.NewE(err)
		}
		teamVpnConfig := fileclient.TeamVpnConfig{
			WGconf:     dev.WireguardConfig.Value,
			DeviceName: dev.Metadata.Name,
			IpAddress:  dev.IPAddress,
		}

		if err := apic.fc.SetVpnTeamConfig(team, &teamVpnConfig); err != nil {
			return nil, fn.NewE(err)
		}
	} else if err != nil {
		return nil, fn.NewE(err)
	}
	if avc == nil {
		avc, err = apic.fc.GetVpnTeamConfig(team)
		if err != nil {
			return nil, fn.NewE(err)
		}
	}
	if avc.WGconf == "" {
		d, err := apic.GetVPNDevice(team, avc.DeviceName)
		if err != nil {
			return nil, fn.NewE(err)
		}

		avc.WGconf = d.WireguardConfig.Value

		if err := apic.fc.SetVpnTeamConfig(team, avc); err != nil {
			return nil, fn.NewE(err)
		}
	}

	return avc, nil
}
