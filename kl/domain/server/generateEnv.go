package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/klbox-docker/devboxfile"
	"github.com/kloudlite/kl/pkg/functions"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type SecretEnv struct {
	Key        string `json:"key"`
	SecretName string `json:"secretName"`
	Value      string `json:"value"`
}

type ConfigEnv struct {
	Key        string `json:"key"`
	ConfigName string `json:"configName"`
	Value      string `json:"value"`
}

type MresEnv struct {
	Key      string `json:"key"`
	MresName string `json:"mresName"`
	Value    string `json:"value"`
}

type EnvRsp struct {
	Secrets []SecretEnv `json:"secrets"`
	Configs []ConfigEnv `json:"configs"`
	Mreses  []MresEnv   `json:"mreses"`
}

type GeneratedEnvs struct {
	EnvVars    map[string]string `json:"envVars"`
	MountFiles map[string]string `json:"mountFiles"`
}

func GenerateEnv() (*GeneratedEnvs, error) {
	klFile, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_generateEnv", map[string]any{
		"klConfig": klFile,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		GeneratedEnvVars GeneratedEnvs `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	// return resp.CoreConfigs, nil
	return &resp.GeneratedEnvVars, nil
}

type Kv struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CSResp map[string]map[string]*Kv
type MountMap map[string]string

func GetLoadMaps() (map[string]string, MountMap, error) {

	kt, err := client.GetKlFile("")
	if err != nil {
		return nil, nil, err
	}

	env, err := EnsureEnv(nil)
	if err != nil {
		return nil, nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, nil, err
	}

	currMreses := kt.EnvVars.GetMreses()
	currSecs := kt.EnvVars.GetSecrets()
	currConfs := kt.EnvVars.GetConfigs()

	currMounts := kt.Mounts.GetMounts()

	respData, err := klFetch("cli_getConfigSecretMap", map[string]any{
		"envName": env.Name,
		"configQueries": func() []any {
			var queries []any
			for _, v := range currConfs {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"configName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range currMounts {
				if fe.Type == client.ConfigType {
					queries = append(queries, map[string]any{
						"configName": fe.Name,
						"key":        fe.Key,
					})
				}
			}

			return queries
		}(),

		"mresQueries": func() []any {
			var queries []any
			for _, rt := range currMreses {
				for _, v := range rt.Env {
					queries = append(queries, map[string]any{
						"mresName": rt.Name,
						"key":      v.RefKey,
					})
				}
			}

			return queries
		}(),

		"secretQueries": func() []any {
			var queries []any
			for _, v := range currSecs {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"secretName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range currMounts {
				if fe.Type == client.SecretType {
					queries = append(queries, map[string]any{
						"secretName": fe.Name,
						"key":        fe.Key,
					})
				}
			}
			return queries
		}(),
	}, &cookie)

	if err != nil {
		return nil, nil, err

	}

	fromResp, err := GetFromResp[EnvRsp](respData)

	if err != nil {
		return nil, nil, err
	}

	result := map[string]string{}

	cmap := CSResp{}

	for _, rt := range currConfs {
		cmap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			cmap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	smap := CSResp{}

	for _, rt := range currSecs {
		smap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			smap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	mmap := CSResp{}
	for _, rt := range currMreses {
		mmap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			mmap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	// ************************[ adding to result|env ]***************************
	for _, v := range fromResp.Configs {
		ent := cmap[v.ConfigName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if cmap[v.ConfigName][v.Key] != nil {
			cmap[v.ConfigName][v.Key].Value = v.Value
		}

	}

	for _, v := range fromResp.Secrets {
		ent := smap[v.SecretName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if smap[v.SecretName][v.Key] != nil {
			smap[v.SecretName][v.Key].Value = v.Value
		}
	}

	for _, v := range fromResp.Mreses {
		ent := mmap[v.MresName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		if mmap[v.MresName][v.Key] != nil {
			mmap[v.MresName][v.Key].Value = v.Value
		}
	}

	// ************************[ handling mounts ]****************************
	mountMap := map[string]string{}

	for _, fe := range currMounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		if fe.Type == client.ConfigType {
			mountMap[pth] = func() string {
				for _, ce := range fromResp.Configs {
					if ce.ConfigName == fe.Name && ce.Key == fe.Key {
						return ce.Value
					}
				}
				return ""
			}()
		} else {
			mountMap[pth] = func() string {
				for _, ce := range fromResp.Secrets {
					if ce.SecretName == fe.Name && ce.Key == fe.Key {
						return ce.Value
					}
				}
				return ""
			}()
		}
	}

	return result, mountMap, nil
}

func syncLockfile(config *devboxfile.DevboxConfig) (map[string]string, error) {
	// check if kl.lock file exists
	_, err := os.Stat("kl.lock")
	packages := map[string]string{}
	if err == nil {
		file, err := os.ReadFile("kl.lock")
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(file, &packages)
		if err != nil {
			return nil, err
		}
	}
	for p := range config.Packages {
		splits := strings.Split(config.Packages[p], "@")
		if len(splits) == 1 {
			splits = append(splits, "latest")
		}
		if _, ok := packages[splits[0]+"@"+splits[1]]; ok {
			continue
		}
		resp, err := http.Get(fmt.Sprintf("https://search.devbox.sh/v1/resolve?name=%s&version=%s", splits[0], splits[1]))
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch package %s", config.Packages[p])
		}
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		type Res struct {
			CommitHash string `json:"commit_hash"`
			Version    string `json:"version"`
		}
		var res Res
		err = json.Unmarshal(all, &res)
		if err != nil {
			return nil, err
		}
		packages[splits[0]+"@"+res.Version] = fmt.Sprintf("nixpkgs/%s#%s", res.CommitHash, splits[0])
	}
	for k, _ := range packages {
		splits := strings.Split(k, "@")
		if !slices.Contains(config.Packages, splits[0]) && !slices.Contains(config.Packages, k) && !slices.Contains(config.Packages, splits[0]+"@latest") {
			delete(packages, k)
		}
	}
	marshal, err := json.Marshal(packages)
	if err != nil {
		return nil, err
	}
	bjson := new(bytes.Buffer)
	if err = json.Indent(bjson, marshal, "", "  "); err != nil {
		return nil, err
	}
	if err = os.WriteFile("kl.lock", bjson.Bytes(), 0644); err != nil {
		return nil, err
	}
	return packages, nil
}

func LoadDevboxConfig() (*devboxfile.DevboxConfig, error) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Suffix = " Syncing Environment..."
	s.Start()
	defer func() {
		s.Stop()
	}()

	envs, mm, err := GetLoadMaps()
	if err != nil {
		return nil, err
	}

	kf, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	// read kl.yml into struct
	klConfig := &devboxfile.DevboxConfig{
		Packages: kf.Packages,
	}

	fm := map[string]string{}

	for _, fe := range kf.Mounts.GetMounts() {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		fm[pth] = mm[pth]
	}

	ev := map[string]string{}
	for k, v := range envs {
		ev[k] = v
	}

	for _, ne := range kf.EnvVars.GetEnvs() {
		ev[ne.Key] = ne.Value
	}

	e, err := client.CurrentEnv()
	if err == nil {
		ev["PURE_PROMPT_SYMBOL"] = fmt.Sprintf("(%s) %s", e.Name, "❯")
	}

	klConfig.Env = ev
	klConfig.KlConfig.Mounts = fm
	lockfile, err := syncLockfile(klConfig)
	klConfig.PackageHashes = lockfile
	return klConfig, nil
}

func MountEnvs(c devboxfile.KlConfig) error {
	for k, v := range c.Mounts {
		if err := os.MkdirAll(filepath.Dir(k), fs.ModePerm); err != nil {
			functions.Warnf("failed to create dir %s", filepath.Dir(k))
			continue
		}

		if err := os.WriteFile(k, []byte(v), fs.ModePerm); err != nil {
			functions.Warnf("failed to write file %s", k)
			continue
		}
	}

	return nil
}
