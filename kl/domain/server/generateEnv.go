package server

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/klbox-docker/devboxfile"
	"github.com/kloudlite/kl/pkg/functions"
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

// this function will fetch real envs from api and return DevboxKlfile with real envs
func LoadDevboxConfig() (*devboxfile.DevboxConfig, error) {
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

	kt, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}

	fm := map[string]string{}

	for _, fe := range kt.Mounts.GetMounts() {
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

	klConfig.Env = ev
	klConfig.KlConfig.Mounts = fm

	// klConfig.KlConfig.InitScripts = kt.InitScripts

	return klConfig, nil
}

func SyncDevboxJsonFile() error {
	if !client.InsideBox() {
		return nil
	}

	kConf, err := LoadDevboxConfig()
	if err != nil {
		return err
	}

	b, err := kConf.ToJson()
	if err != nil {
		return err
	}

	if err := os.WriteFile(client.DEVBOX_JSON_PATH, b, os.ModePerm); err != nil {
		return err
	}

	if err := MountEnvs(kConf.KlConfig); err != nil {
		return err
	}

	return nil
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
