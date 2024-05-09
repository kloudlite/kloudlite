package server

import (
	"encoding/json"

	"github.com/kloudlite/kl/domain/client"
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

	cookie, err := getCookie()
	if err != nil {
		return nil, nil, err
	}

	respData, err := klFetch("cli_getConfigSecretMap", map[string]any{
		"envName": env.Name,
		"configQueries": func() []any {
			var queries []any
			for _, v := range kt.Configs {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"configName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range kt.FileMount.Mounts {
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
			for _, rt := range kt.Mres {
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
			for _, v := range kt.Secrets {
				for _, vv := range v.Env {
					queries = append(queries, map[string]any{
						"secretName": v.Name,
						"key":        vv.RefKey,
					})
				}
			}

			for _, fe := range kt.FileMount.Mounts {
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

	for _, rt := range kt.Configs {
		cmap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			cmap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	smap := CSResp{}

	for _, rt := range kt.Secrets {
		smap[rt.Name] = map[string]*Kv{}
		for _, v := range rt.Env {
			smap[rt.Name][v.RefKey] = &Kv{
				Key: v.Key,
			}
		}
	}

	mmap := CSResp{}
	for _, rt := range kt.Mres {
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

	for _, fe := range kt.FileMount.Mounts {
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
