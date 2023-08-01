package helm

import (
	"context"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/time"

	"github.com/kloudlite/operator/pkg/logging"
	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/client-go/rest"
)

type hClient struct {
	hc        *helmclient.HelmClient
	defaultNs string
	logger    logging.Logger
}

func getRelease(hc *helmclient.HelmClient, namespace string, releaseName string) (*release.Release, error) {
	hc.Settings.SetNamespace(namespace)

	r, err := hc.GetRelease(releaseName)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}

func isReleaseBeingDeleted(release *release.Release) bool {
	return release != nil && release.Info != nil && release.Info.Deleted.Sub(time.Time{}) > 0
}

func (c *hClient) UninstallRelease(ctx context.Context, namespace string, releaseName string) error {
	c.hc.Settings.SetNamespace(namespace)

	release, err := getRelease(c.hc, namespace, releaseName)
	if err != nil {
		return nil
	}
	if release == nil {
		return nil
	}
	return c.hc.UninstallReleaseByName(releaseName)
}

func (c *hClient) HasBeenDeleted(ctx context.Context, namespace string, releaseName string) (bool, error) {
	c.hc.Settings.SetNamespace(namespace)

	r, err := c.hc.GetRelease(releaseName)
	if err != nil {
		return false, err
	}
	if r.Info == nil {
		return false, err
	}

	return r.Info.Deleted.Sub(time.Time{}) > 0, nil
}

func (c *hClient) GetRelease(ctx context.Context, namespace string, releaseName string) (*release.Release, error) {
	c.hc.Settings.SetNamespace(namespace)
	return c.hc.GetRelease(releaseName)
}

func (c *hClient) AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := c.hc.AddOrUpdateChartRepo(repo.Entry{
		Name: entry.Name,
		URL:  entry.Url,
	}); err != nil {
		return err
	}
	return nil
}

func (c *hClient) InstallOrUpgradeChart(ctx context.Context, namespace string, spec ChartSpec, opts UpgradeOpts) (*release.Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.hc.Settings.SetNamespace(namespace)

	if spec.Namespace != namespace {
		spec.Namespace = namespace
	}

	release, err := getRelease(c.hc, namespace, spec.ReleaseName)
	if err != nil {
		return nil, err
	}

	if isReleaseBeingDeleted(release) {
		return nil, fmt.Errorf("release with name %q is currently being deleted, can not perform install/upgrade now, try after some time", spec.ReleaseName)
	}

	if release != nil && opts.UpgradeOnlyIfValuesChanged {
		helmValues, err := c.hc.GetReleaseValues(spec.ReleaseName, false)
		if err != nil {
			return nil, err
		}

		if areHelmValuesEqual(helmValues, []byte(spec.ValuesYaml)) {
			if c.logger != nil {
				c.logger.Infof("chart release with name %q, already exists with same values. skipping upgrade (as opts UpgradeOnlyIfValuesChanged=true)", spec.ReleaseName)
			}
			return release, nil
		}
	}

	return c.hc.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
		ReleaseName: spec.ReleaseName,
		ChartName:   spec.ChartName,
		Namespace:   spec.Namespace,
		Version:     spec.Version,
		ValuesYaml:  spec.ValuesYaml,
	}, &helmclient.GenericHelmOptions{})
}

func (c *hClient) GetReleaseValues(ctx context.Context, namespace string, releaseName string) (map[string]any, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.hc.Settings.SetNamespace(namespace)
	return c.hc.GetReleaseValues(releaseName, false)
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
		Debug:            true,
	})
	if err != nil {
		return nil, err
	}

	return &hClient{
		hc:     hc,
		logger: opts.Logger,
	}, nil
}

func NewHelmClientOrDie(config *rest.Config, opts ClientOptions) Client {
	c, err := NewHelmClient(config, opts)
	if err != nil {
		panic(err)
	}
	return c
}
