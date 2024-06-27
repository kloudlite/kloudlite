package packages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

type PackageInfo struct {
	ID           int      `json:"id"`
	CommitHash   string   `json:"commit_hash"`
	System       string   `json:"system"`
	LastUpdated  int      `json:"last_updated"`
	StoreHash    string   `json:"store_hash"`
	StoreName    string   `json:"store_name"`
	StoreVersion string   `json:"store_version"`
	MetaName     string   `json:"meta_name"`
	MetaVersion  []string `json:"meta_version"`
	AttrPaths    []string `json:"attr_paths"`
	Version      string   `json:"version"`
	Summary      string   `json:"summary"`
}

type PackageVersion struct {
	PackageInfo

	Name    string                 `json:"name"`
	Systems map[string]PackageInfo `json:"systems,omitempty"`
}

type Package struct {
	Name        string           `json:"name"`
	NumVersions int              `json:"num_versions"`
	Versions    []PackageVersion `json:"versions,omitempty"`
}

type SearchResults struct {
	NumResults int       `json:"num_results"`
	Packages   []Package `json:"packages,omitempty"`
}

const searchAPIEndpoint = "https://search.devbox.sh"

func Search(ctx context.Context, query string) (*SearchResults, error) {
	if query == "" {
		return nil, fmt.Errorf("query should not be empty")
	}
	defer spinner.Client.UpdateMessage(fmt.Sprintf("searching for package %s", query))()

	endpoint, err := url.JoinPath(searchAPIEndpoint, "v1/search")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("package %s not found", query)
		}

		return nil, fn.NewE(err)
	}
	url := endpoint + "?q=" + url.QueryEscape(query)

	return caller[SearchResults](ctx, url)
}

func Resolve(ctx context.Context, name string) (string, error) {
	if !strings.Contains(name, "@") {
		sr, err := Search(ctx, name)
		if err != nil {
			return "", fn.NewE(err)
		}

		pkg, err := fzf.FindOne(sr.Packages, func(item Package) string {
			return item.Name
		}, fzf.WithPrompt("select a package"))

		if err != nil {
			return "", fn.NewE(err)
		}

		version, err := fzf.FindOne(pkg.Versions, func(item PackageVersion) string {
			return fmt.Sprintf("%s %s", item.Version, item.Summary)
		}, fzf.WithPrompt("select a version"))

		if err != nil {
			return "", fn.NewE(err)
		}

		return fmt.Sprintf("%s@%s", pkg.Name, version.Version), nil
	}

	splits := strings.Split(name, "@")

	if strings.TrimSpace(splits[0]) == "" || strings.TrimSpace(splits[1]) == "" {
		return "", fmt.Errorf("package %s is invalid", name)
	}

	type Res struct {
		CommitHash string `json:"commit_hash"`
		Version    string `json:"version"`
	}

	sr, err := caller[Res](ctx, fmt.Sprintf("%s/v1/resolve?name=%s&version=%s", searchAPIEndpoint, splits[0], splits[1]))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", fmt.Errorf("package %s not found", name)
		}
		return "", fn.NewE(err)
	}

	return fmt.Sprintf("%s@%s", splits[0], sr.Version), nil
}

var ErrNotFound = fn.Error("not found")

func caller[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fn.Errorf("GET %s: %w", url, err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fn.Errorf("GET %s: %w", url, err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fn.Errorf("GET %s: read respoonse body: %w", url, err)
	}

	if response.StatusCode == 404 {
		return nil, ErrNotFound
	}

	if response.StatusCode >= 400 {
		return nil, fn.Errorf("GET %s: unexpected status code %s: %s",
			url,
			response.Status,
			data,
		)
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fn.Errorf("GET %s: unmarshal response JSON: %w", url, err)
	}
	return &result, nil
}
