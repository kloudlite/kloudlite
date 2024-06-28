package fileclient

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type AccountVpnConfig struct {
	WGconf     string `json:"wg"`
	DeviceName string `json:"device"`
}

func (a *AccountVpnConfig) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func (a *AccountVpnConfig) Unmarshal(b []byte) error {
	return json.Unmarshal(b, a)
}

func (c *fclient) GetVpnAccountConfig(account string) (*AccountVpnConfig, error) {
	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "vpn"), 0755); err != nil {
		return nil, fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "vpn", fmt.Sprintf("%s.json", account))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, err
	}

	var accVPNConfig AccountVpnConfig
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fn.NewE(err, "failed to read vpn config")
	}

	if err := accVPNConfig.Unmarshal(b); err != nil {
		return nil, fn.NewE(err, "failed to parse vpn config")
	}

	return &accVPNConfig, nil
}

func (c *fclient) SetVpnAccountConfig(account string, avc *AccountVpnConfig) error {
	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "vpn"), 0755); err != nil {
		return fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "vpn", fmt.Sprintf("%s.json", account))

	marshal, err := avc.Marshal()
	if err != nil {
		return fn.NewE(err)
	}
	err = os.WriteFile(cfgPath, marshal, 0644)
	if err != nil {
		return fn.NewE(err)
	}

	return nil
}
