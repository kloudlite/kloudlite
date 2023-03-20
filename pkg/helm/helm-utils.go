package helm

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/time"

	"github.com/kloudlite/operator/pkg/errors"
	"github.com/kloudlite/operator/pkg/logging"
	helmclient "github.com/mittwald/go-helm-client"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

type ClientOptions struct {
	Namespace string
	Logger    logging.Logger
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
	Name              string
	Url               string
	LocalChartArchive string
}

type Client interface {
	AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error
	InstallOrUpgradeChart(ctx context.Context, spec ChartSpec) (*release.Release, error)
	GetReleaseValues(ctx context.Context, releaseName string) (map[string]any, error)
	GetRelease(ctx context.Context, releaseName string) (*release.Release, error)
	HasBeenDeleted(ctx context.Context, releaseName string) (bool, error)
	UninstallRelease(ctx context.Context, releaseName string) error

	EnsureRelease(ctx context.Context, spec ChartSpec) (*release.Release, error)
}

type hClient struct {
	hc         helmclient.Client
	restConfig *rest.Config
	defaultNs  string
	logger     logging.Logger
}

func getRelease(hc helmclient.Client, releaseName string) (*release.Release, error) {
	r, err := hc.GetRelease(releaseName)
	if err != nil {
		if strings.Contains(err.Error(), "release: not found") {
			return nil, nil
		}
		return nil, err
	}
	return r, nil
}

func (c *hClient) EnsureRelease(ctx context.Context, spec ChartSpec) (*release.Release, error) {
	hc, err := func() (helmclient.Client, error) {
		if spec.Namespace == c.defaultNs {
			return c.hc, nil
		}
		return newHelmClient(c.restConfig, &helmclient.Options{Namespace: spec.Namespace})
	}()

	if err != nil {
		return nil, err
	}

	release, err := getRelease(hc, spec.ReleaseName)
	if err != nil {
		return nil, err
	}

	if release != nil && release.Info != nil && (release.Info.Deleted.Sub(time.Time{}) > 0 || release.Info.Status.String() == "failed") {
		if c.logger != nil {
			c.logger.Infof("helm release has been deleted or failed: %s, uninstalling now ...", spec.ReleaseName)
		}
		if err := hc.UninstallReleaseByName(spec.ReleaseName); err != nil {
			return nil, errors.NewEf(err, "uninstalling release %s", spec.ReleaseName)
		}
	}

	helmValues, _ := hc.GetReleaseValues(spec.ReleaseName, false)
	if areHelmValuesEqual(helmValues, []byte(spec.ValuesYaml)) {
		return nil, nil
	}

	return hc.InstallOrUpgradeChart(ctx, &helmclient.ChartSpec{
		ReleaseName: spec.ReleaseName,
		ChartName:   spec.ChartName,
		Namespace:   spec.Namespace,
		ValuesYaml:  spec.ValuesYaml,
	}, &helmclient.GenericHelmOptions{})
}

// UninstallRelease implements Client
func (c *hClient) UninstallRelease(ctx context.Context, releaseName string) error {
	return c.hc.UninstallReleaseByName(releaseName)
}

func (c *hClient) HasBeenDeleted(ctx context.Context, releaseName string) (bool, error) {
	r, err := c.hc.GetRelease(releaseName)
	if err != nil {
		return false, err
	}
	if r.Info == nil {
		return false, err
	}
	return r.Info.Deleted.Sub(time.Time{}) > 0, nil
}

func (c *hClient) GetRelease(ctx context.Context, releaseName string) (*release.Release, error) {
	return c.hc.GetRelease(releaseName)
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

func newHelmClient(config *rest.Config, opts *helmclient.Options) (helmclient.Client, error) {
	return helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options:    opts,
		RestConfig: config,
	})
}

func NewHelmClient(config *rest.Config, opts ClientOptions) (Client, error) {
	hc, err := newHelmClient(config, &helmclient.Options{Namespace: opts.Namespace})
	if err != nil {
		return nil, err
	}

	return &hClient{hc: hc, restConfig: config, defaultNs: opts.Namespace, logger: opts.Logger}, nil
}

func NewHelmClientOrDie(config *rest.Config, opts ClientOptions) Client {
	c, err := NewHelmClient(config, opts)
	if err != nil {
		panic(err)
	}
	return c
}

func areHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	b, err := json.Marshal(releaseValues)
	if err != nil {
		return false
	}

	tv, err := yaml.YAMLToJSON(templateValues)
	if err != nil {
		return false
	}

	if len(b) != len(tv) || !bytes.Equal(b, tv) {
		return false
	}
	return true
}

func AreHelmValuesEqual(releaseValues map[string]any, templateValues []byte) bool {
	return areHelmValuesEqual(releaseValues, templateValues)
}
