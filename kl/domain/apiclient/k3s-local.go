package apiclient

import (
	"os"

	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Cluster struct {
	ClusterToken   string          `json:"clusterToken"`
	Name           string          `json:"name"`
	InstallCommand *InstallCommand `json:"installCommand"`
	Metadata       struct {
		Name   string            `json:"name"`
		Labels map[string]string `json:"labels"`
	} `json:"metadata"`
}

type InstallCommand struct {
	ChartRepo    string `json:"chart-repo"`
	ChartVersion string `json:"chart-version"`
	CRDsURL      string `json:"crds-url"`
	HelmValues   struct {
		AccountName           string `json:"accountName"`
		ClusterName           string `json:"clusterName"`
		ClusterToken          string `json:"clusterToken"`
		KloudliteDNSSuffix    string `json:"kloudliteDNSSuffix"`
		MessageOfficeGRPCAddr string `json:"messageOfficeGRPCAddr"`
	} `json:"helm-values"`
}

func (apic *apiClient) getClustersOfAccount(account string) ([]Cluster, error) {
	cookie, err := getCookie(fn.MakeOption("accountName", account))
	if err != nil {
		return nil, fn.NewE(err)
	}
	fetch, err := klFetch("cli_listAccountClusters", map[string]any{}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	clusters, err := GetFromRespForEdge[Cluster](fetch)
	if err != nil {
		return nil, fn.NewE(err)
	}
	return clusters, nil
}

func (apic *apiClient) GetClusterConfig(account string) (*fileclient.AccountClusterConfig, error) {

	existingClusters, err := apic.getClustersOfAccount(account)
	if err != nil {
		return nil, err
	}
	var selectedCluster *Cluster
	wgconfig, err := apic.fc.GetWGConfig()
	if err != nil {
		return nil, err
	}
	for _, c := range existingClusters {
		if c.Metadata.Labels["kloudlite.io/local-uuid"] == wgconfig.UUID {
			selectedCluster = &c
			err := apic.enrichClusterWithInstructions(account, selectedCluster)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if selectedCluster == nil {
		selectedCluster, err = apic.createClusterForAccount(account)
		if err != nil {
			return nil, fn.NewE(err)
		}
	}

	config := fileclient.AccountClusterConfig{
		ClusterToken: selectedCluster.ClusterToken,
		ClusterName:  selectedCluster.Metadata.Name,
		InstallCommand: fileclient.InstallCommand{
			ChartRepo:    selectedCluster.InstallCommand.ChartRepo,
			ChartVersion: selectedCluster.InstallCommand.ChartVersion,
			CRDsURL:      selectedCluster.InstallCommand.CRDsURL,
			HelmValues: fileclient.InstallHelmValues{
				AccountName:           selectedCluster.InstallCommand.HelmValues.AccountName,
				ClusterName:           selectedCluster.InstallCommand.HelmValues.ClusterName,
				ClusterToken:          selectedCluster.InstallCommand.HelmValues.ClusterToken,
				KloudliteDNSSuffix:    selectedCluster.InstallCommand.HelmValues.KloudliteDNSSuffix,
				MessageOfficeGRPCAddr: selectedCluster.InstallCommand.HelmValues.MessageOfficeGRPCAddr,
			},
		},
	}
	config.WGConfig = *wgconfig
	err = apic.fc.SetClusterConfig(account, &config)
	if err != nil {
		return nil, fn.NewE(err)
	}
	return &config, nil
}

func getClusterName(clusterName, account string) (*CheckName, error) {
	cookie, err := getCookie(fn.MakeOption("accountName", account))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_infraCheckNameAvailability", map[string]any{
		"resType": ClusterType,
		"name":    clusterName,
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

func (apic *apiClient) enrichClusterWithInstructions(account string, d *Cluster) error {
	cookie, err := getCookie(fn.MakeOption("accountName", account))
	if err != nil {
		return fn.NewE(err)
	}

	respData, err := klFetch("cli_clusterReferenceInstructions", map[string]any{
		"name": d.Metadata.Name,
	}, &cookie)

	if err != nil {
		return fn.NewE(err)
	}

	instruction, err := GetFromResp[InstallCommand](respData)
	if err != nil {
		return fn.NewE(err)
	}

	d.InstallCommand = instruction
	return nil
}

func (apic *apiClient) createCluster(hostName, account string) (*Cluster, error) {
	user, err := apic.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	userName := user.Name + "-" + hostName
	cn, err := getClusterName(userName, account)
	if err != nil {
		return nil, fn.NewE(err)
	}

	cookie, err := getCookie(fn.MakeOption("accountName", account))
	if err != nil {
		return nil, fn.NewE(err)
	}

	dn := userName
	if !cn.Result {
		if len(cn.SuggestedNames) == 0 {
			return nil, fn.Errorf("no suggested names for cluster %s", userName)
		}
		dn = cn.SuggestedNames[0]
	}

	wgconfig, err := apic.fc.GetWGConfig()
	if err != nil {
		return nil, err
	}

	fn.Logf("creating new cluster %s\n", dn)
	respData, err := klFetch("cli_createClusterReference", map[string]any{
		"cluster": map[string]any{
			"metadata": map[string]any{
				"name": dn,
				"labels": map[string]string{
					"kloudlite.io/k3scluster": "true",
					"kloudlite.io/local-uuid": wgconfig.UUID,
					"kloudlite.io/owned-by":   user.UserId,
				},
			},
			"displayName": userName,
			"visibility":  map[string]string{"mode": "private"},
		},
	}, &cookie)
	if err != nil {
		return nil, fn.Errorf("failed to create vpn: %s", err.Error())
	}
	d, err := GetFromResp[Cluster](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	err = apic.enrichClusterWithInstructions(account, d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (apic *apiClient) createClusterForAccount(account string) (*Cluster, error) {
	hostName, err := os.Hostname()
	if err != nil {
		return nil, fn.NewE(err)
	}
	cluster, err := apic.createCluster(hostName, account)
	if err != nil {
		return nil, fn.NewE(err)
	}
	return cluster, nil
}
