package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"

	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type ClientOptions struct {
	Namespace string
}

// ChartSpec is subset of type [helmclient](github.com/mittwald/go-helm-client).ChartSpec
type ChartSpec struct {
	ReleaseName string
	ChartName   string
	Namespace   string
	ValuesYaml  string
}

// RepoEntry is subset of type [repo](helm.sh/helm/v3/pkg/repo).Entry
type RepoEntry struct {
	Name string
	Url  string
}

type Client interface {
	AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error
	InstallOrUpgradeChart(ctx context.Context, spec ChartSpec) (*release.Release, error)
	GetReleaseValues(ctx context.Context, releaseName string) (map[string]any, error)
}

type hClient struct {
	hc helmclient.Client
}

func (c *hClient) AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return c.hc.AddOrUpdateChartRepo(repo.Entry{
		Name: entry.Name,
		URL:  entry.Url,
	})
}

func (c *hClient) InstallOrUpgradeChart(ctx context.Context, spec ChartSpec) (*release.Release, error) {
	return c.hc.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
		ReleaseName: spec.ReleaseName,
		ChartName:   spec.ChartName,
		Namespace:   spec.Namespace,
		ValuesYaml:  spec.ValuesYaml,
	}, &helmclient.GenericHelmOptions{})
}

func (c *hClient) GetReleaseValues(ctx context.Context, releaseName string) (map[string]any, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return c.hc.GetReleaseValues(releaseName, false)
}

func NewHelmClient(config *rest.Config, opts ClientOptions) (Client, error) {
	hc, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: opts.Namespace,
		},
		RestConfig: config,
	})
	if err != nil {
		return nil, err
	}
	return &hClient{hc: hc}, nil
}

func AreHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	b, err := json.Marshal(releaseValues)
	if err != nil {
		return false
	}

	tv, err := yaml.YAMLToJSON(templateValues)
	if err != nil {
		return false
	}

	if len(b) != len(tv) || bytes.Compare(b, tv) != 0 {
		return false
	}
	return true
}
