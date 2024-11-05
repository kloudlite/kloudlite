package packages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
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
		return nil, fn.Errorf("query should not be empty")
	}
	defer spinner.Client.UpdateMessage(fmt.Sprintf("searching for package %s", query))()

	endpoint, err := url.JoinPath(searchAPIEndpoint, "v1/search")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, fn.Errorf("package %s not found", query)
		}

		return nil, fn.NewE(err)
	}
	url := endpoint + "?q=" + url.QueryEscape(query)

	return caller[SearchResults](ctx, url)
}

func Resolve(ctx context.Context, pname string) (string, string, error) {
	var name string
	var v string
	if !strings.Contains(pname, "@") {
		sr, err := Search(ctx, pname)
		if err != nil {
			return "", "", fn.NewE(err)
		}

		pkg, err := fzf.FindOne(sr.Packages, func(item Package) string {
			return item.Name
		}, fzf.WithPrompt("select a package"))

		if err != nil {
			return "", "", fn.NewE(err)
		}

		version, err := fzf.FindOne(pkg.Versions, func(item PackageVersion) string {
			return fmt.Sprintf("%s %s", item.Version, item.Summary)
		}, fzf.WithPrompt("select a version"))

		if err != nil {
			return "", "", fn.NewE(err)
		}
		name = version.Name
		v = version.Version
	} else {
		splits := strings.Split(name, "@")

		if strings.TrimSpace(splits[0]) == "" || strings.TrimSpace(splits[1]) == "" {
			return "", "", fn.Errorf("package %s is invalid", name)
		}
		name = splits[0]
		v = splits[1]
	}

	type System struct {
		AttrPaths []string `json:"attr_paths"`
	}

	type Res struct {
		CommitHash string            `json:"commit_hash"`
		Version    string            `json:"version"`
		Systems    map[string]System `json:"systems"`
	}

	platform := os.Getenv("PLATFORM_ARCH") + "-linux"
	sr, err := caller[Res](ctx, fmt.Sprintf("%s/v1/resolve?name=%s&version=%s&platform=%s", searchAPIEndpoint, name, v, platform))

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", "", fn.Errorf("package %s not found", name)
		}
		return "", "", fn.NewE(err)
	}

	return fmt.Sprintf("%s@%s", name, sr.Version), fmt.Sprintf("%s#%s", sr.CommitHash, sr.Systems[platform].AttrPaths[0]), nil
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
