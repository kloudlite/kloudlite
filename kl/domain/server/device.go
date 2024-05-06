package server

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"net/http"
	"time"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type DevicePort struct {
	Port       int `json:"port"`
	TargetPort int `json:"targetPort,omitempty"`
}

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
		Generation        string            `json:"generation"`
		Labels            map[string]string `json:"labels"`
		Name              string            `json:"name"`
		Namespace         string            `json:"namespace"`
	} `json:"metadata"`
	PrivateKey      string `json:"privateKey"`
	PublicEndpoint  string `json:"publicEndpoint"`
	PublicKey       string `json:"publicKey"`
	RecordVersion   string `json:"recordVersion"`
	UpdateTime      string `json:"updateTime"`
	WireguardConfig struct {
		Value    string `json:"value"`
		Encoding string `json:"encoding"`
	} `json:"wireguardConfig,omitempty"`
}

const (
	VPNDEVICEGVPN = "default"
)

type DeviceList struct {
	Edges Edges[Env] `json:"edges"`
}

func GetVPNDevice(devName string, options ...fn.Option) (*Device, error) {
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
		//"envName":    envName,
		"gvpn":       VPNDEVICEGVPN,
		"deviceName": devName,
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

func ListVPNDevice(options ...fn.Option) ([]Device, error) {
	accountName := fn.GetOption(options, "accountName")

	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getGlobalVpnDevices", map[string]any{
		"gvpn": VPNDEVICEGVPN,
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)
	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Device](respData); err != nil {
		return nil, err
	} else {
		if len(fromResp) < 1 {
			return nil, errors.New("No Global VPN devices found. Please create one from dashboard.")
		}
		return fromResp, nil
	}
}

func SelectDevice(devName string, options ...fn.Option) (*Device, error) {
	persistSelectedDevice := func(devName string) error {
		err := client.SelectDevice(devName)
		if err != nil {
			return err
		}
		return nil
	}

	devices, err := ListVPNDevice(options...)
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

	dev, err := fzf.FindOne(
		devices,
		func(dev Device) string {
			return fmt.Sprintf("%s (%s)", dev.DisplayName, dev.Metadata.Name)
		},
		fzf.WithPrompt("Select Environment > "),
	)

	if err != nil {
		return nil, err
	}

	if err := persistSelectedDevice(dev.Metadata.Name); err != nil {
		return nil, err
	}

	return dev, nil
}

func CheckDeviceStatus() bool {

	httpClient := http.Client{Timeout: 200 * time.Millisecond}
	if _, err := httpClient.Get("http://10.13.0.1:17171"); err != nil {
		return false
	}

	return true
}
