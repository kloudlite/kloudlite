package helm

import (
	"context"
	"fmt"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"

	"k8s.io/client-go/rest"
)

type client struct {
	hc helmclient.Client
}

func (c *client) AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := c.hc.AddOrUpdateChartRepo(repo.Entry{
		Name: entry.Name,
		URL:  entry.URL,
	}); err != nil {
		return err
	}
	return nil
}

func (c *client) TemplateChart(ctx context.Context, chart *ChartSpec) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return c.hc.TemplateChart(&helmclient.ChartSpec{
		ReleaseName: chart.ReleaseName,
		Namespace:   chart.Namespace,
		ChartName:   chart.ChartName,
		Version:     chart.Version,
		ValuesYaml:  chart.ValuesYaml,
	}, nil)
}

func newHelmClient(config *rest.Config, opts *helmclient.Options) (*helmclient.HelmClient, error) {
	hc, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options:    opts,
		RestConfig: config,
	})
	if err != nil {
		return nil, err
	}

	h, ok := hc.(*helmclient.HelmClient)
	if !ok {
		return nil, fmt.Errorf("unexpected helmclient type: %T, should have been of type: *helmclient.HelmClient", hc)
	}
	return h, nil
}

func NewHelmClient(config *rest.Config, opts ClientOptions) (Client, error) {
	hc, err := newHelmClient(config, &helmclient.Options{
		RepositoryCache:  opts.RepositoryCacheDir,
		RepositoryConfig: opts.RepositoryConfigFile,
		Debug:            false,
	})
	if err != nil {
		return nil, err
	}

	return &client{hc: hc}, nil
}
