package helm

import (
	"context"
)

type ClientOptions struct {
	// path to the local helm repository cache, which keeps all index.yaml, tar.gz archives
	RepositoryCacheDir string
	// path to the local helm repository config, which is just an index.yaml containing entries for all charts this client has encountered
	RepositoryConfigFile string
}

// ChartSpec is subset of type [helmclient.ChartSpec](https://pkg.go.dev/github.com/nxtcoder17/go-helm-client#ChartSpec)
type ChartSpec struct {
	RepoName    string
	ReleaseName string
	Namespace   string

	ChartName  string
	Version    string
	ValuesYaml string
}

// RepoEntry is subset of type [type](https://pkg.go.dev/helm.sh/helm/v3/pkg/repo#Entry)
type RepoEntry struct {
	Name string
	URL  string
}

type Client interface {
	AddOrUpdateChartRepo(ctx context.Context, entry RepoEntry) error
	TemplateChart(ctx context.Context, chart *ChartSpec) ([]byte, error)
}
