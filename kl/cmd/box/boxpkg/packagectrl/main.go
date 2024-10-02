package packagectrl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/fjson"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Packages map[string]string

func (p *Packages) Marshal() ([]byte, error) {
	return fjson.Marshal(p)
}

func (p *Packages) Unmarshal(b []byte) error {
	return fjson.Unmarshal(b, p)
}

func SyncLockfileWithNewConfig(config fileclient.KLFileType) (map[string]string, error) {
	_, err := os.Stat("kl.lock")
	packages := Packages{}
	if err == nil {
		file, err := os.ReadFile("kl.lock")
		if err != nil {
			return nil, fn.NewE(err)
		}

		if err := packages.Unmarshal(file); err != nil {
			return nil, fn.NewE(err)
		}
	}

	packagesMap := make(map[string]string)
	for k := range packages {
		splits := strings.Split(k, "@")
		if len(splits) != 2 {
			continue
		}

		packagesMap[splits[0]] = splits[1]
	}

	for p := range config.Packages {
		splits := strings.Split(config.Packages[p], "@")
		if len(splits) == 1 {
			if _, ok := packagesMap[splits[0]]; ok {
				continue
			}

			splits = append(splits, "latest")
		}

		if _, ok := packages[splits[0]+"@"+splits[1]]; ok {
			continue
		}

		platform := os.Getenv("PLATFORM_ARCH") + "-linux"
		if platform == "-linux" {
			platform = "x86_64-linux"
		}

		resp, err := http.Get(fmt.Sprintf("https://search.devbox.sh/v1/resolve?name=%s&version=%s&platform=%s", splits[0], splits[1], platform))
		if err != nil {
			return nil, fn.NewE(err)
		}

		if resp.StatusCode != 200 {
			return nil, fn.Errorf("failed to fetch package %s", config.Packages[p])
		}

		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fn.NewE(err)
		}

		type System struct {
			AttrPaths []string `json:"attr_paths"`
		}

		type Res struct {
			CommitHash string            `json:"commit_hash"`
			Version    string            `json:"version"`
			Systems    map[string]System `json:"systems"`
		}

		var res Res
		err = json.Unmarshal(all, &res)
		if err != nil {
			return nil, fn.NewE(err)
		}

		packages[splits[0]+"@"+res.Version] = fmt.Sprintf("nixpkgs/%s#%s", res.CommitHash, res.Systems[platform].AttrPaths[0])
	}

	for k := range packages {
		splits := strings.Split(k, "@")
		if !slices.Contains(config.Packages, splits[0]) && !slices.Contains(config.Packages, k) && !slices.Contains(config.Packages, splits[0]+"@latest") {
			delete(packages, k)
		}
	}

	marshal, err := packages.Marshal()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if err = os.WriteFile("kl.lock", marshal, 0644); err != nil {
		return nil, fn.NewE(err)
	}

	return packages, nil
}
