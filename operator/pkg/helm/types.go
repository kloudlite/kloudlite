package helm

import (
	"context"

	"github.com/kloudlite/operator/pkg/logging"
	"helm.sh/helm/v3/pkg/release"
)

type ClientOptions struct {
	// path to the local helm repository cache, which keeps all index.yaml, tar.gz archives
	RepositoryCacheDir string
	// path to the local helm repository config, which is just an index.yaml containing entries for all charts this client has encountered
	RepositoryConfigFile string
	Logger               logging.Logger
}

// ChartSpec is subset of type [helmclient.ChartSpec](https://pkg.go.dev/github.com/nxtcoder17/go-helm-client#ChartSpec)
type ChartSpec struct {
	ReleaseName string
	Namespace   string

	ChartName  string
	Version    string
	ValuesYaml string
}

// RepoEntry is subset of type [type](https://pkg.go.dev/helm.sh/helm/v3/pkg/repo#Entry)
type RepoEntry struct {
	Name              string
	Url               string
	LocalChartArchive string
}

type UpgradeOpts struct {
	UpgradeOnlyIfValuesChanged bool
}

type Client interface {
	// chart repo
	AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error
	// HasChartRepo(ctx context.Context, name string) (bool, error)

	// release
	GetReleaseValues(ctx context.Context, namespace string, releaseName string) (map[string]any, error)
	GetRelease(ctx context.Context, namespace string, releaseName string) (*release.Release, error)
	HasBeenDeleted(ctx context.Context, namespace string, releaseName string) (bool, error)
	UninstallRelease(ctx context.Context, namespace string, releaseName string) error

	// install or upgrade release based on a chart
	//EnsureRelease(ctx context.Context, namespace string, spec ChartSpec) (*release.Release, error)
	InstallOrUpgradeChart(ctx context.Context, namespace string, spec ChartSpec, opts UpgradeOpts) (*release.Release, error)

	GetLastOperationLogs() string
}
