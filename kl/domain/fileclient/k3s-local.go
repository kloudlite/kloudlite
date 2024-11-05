package fileclient

import (
	"encoding/json"
	"errors"
	"fmt"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"os"
	"path"
)

type TeamClusterConfig struct {
	ClusterToken   string `json:"clusterToken"`
	ClusterName    string `json:"cluster"`
	InstallCommand InstallCommand
	Installed      bool
	WGConfig       WGConfig
	Version        string
	GatewayIP      string
	ClusterCIDR    string
}

type InstallHelmValues struct {
	TeamName              string `json:"accountName"`
	ClusterName           string `json:"clusterName"`
	ClusterToken          string `json:"clusterToken"`
	KloudliteDNSSuffix    string `json:"kloudliteDNSSuffix"`
	MessageOfficeGRPCAddr string `json:"messageOfficeGRPCAddr"`
}

type InstallCommand struct {
	ChartRepo    string `json:"chart-repo"`
	ChartVersion string `json:"chart-version"`
	CRDsURL      string `json:"crds-url"`
	HelmValues   InstallHelmValues
}

func (a *TeamClusterConfig) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func (a *TeamClusterConfig) Unmarshal(b []byte) error {
	return json.Unmarshal(b, a)
}

func (c *fclient) GetClusterConfig(team string) (*TeamClusterConfig, error) {

	if team == "" {
		return nil, fn.Error("team is required")
	}

	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "k3s-local"), 0755); err != nil {
		return nil, fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "k3s-local", fmt.Sprintf("%s.json", team))
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, err
	}

	var accClusterConfig TeamClusterConfig
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fn.NewE(err, "failed to read k3s-local config")
	}

	if err := accClusterConfig.Unmarshal(b); err != nil {
		return nil, fn.NewE(err, "failed to parse k3s-local config")
	}

	wgconf, err := c.GetWGConfig()
	if err != nil {
		return nil, fn.NewE(err)
	}
	accClusterConfig.WGConfig = *wgconf
	return &accClusterConfig, nil
}

func (c *fclient) SetClusterConfig(team string, accClusterConfig *TeamClusterConfig) error {
	if team == "" {
		return fn.Error("team is required")
	}

	cfgFolder := c.configPath

	if err := os.MkdirAll(path.Join(cfgFolder, "k3s-local"), 0755); err != nil {
		return fn.NewE(err)
	}

	cfgPath := path.Join(cfgFolder, "k3s-local", fmt.Sprintf("%s.json", team))

	marshal, err := accClusterConfig.Marshal()
	if err != nil {
		return fn.NewE(err)
	}
	err = os.WriteFile(cfgPath, marshal, 0644)
	if err != nil {
		return fn.NewE(err)
	}

	return nil
}

func (c *fclient) DeleteClusterData(team string) error {
	defer spinner.Client.UpdateMessage("removing cluster data")()
	cfgFolder := c.configPath

	cfgPath := path.Join(cfgFolder, "k3s-local", fmt.Sprintf("%s.json", team))

	_, err := os.Stat(cfgPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fn.NewE(err)
	}

	err = os.Remove(cfgPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fn.NewE(err)
	}

	return nil
}
