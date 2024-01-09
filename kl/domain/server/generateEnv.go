package server

import (
	// "encoding/json"
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

type EnvRsp struct {
	Secrets []SecretEnv `json:"secrets"`
	Configs []ConfigEnv `json:"configs"`
}

type GeneratedEnvs struct {
	EnvVars    map[string]string `json:"envVars"`
	MountFiles map[string]string `json:"mountFiles"`
}

func GenerateEnv() (*GeneratedEnvs, error) {
	klFile, err := client.GetKlFile(nil)
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

	projectId, err := client.CurrentProjectName()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_generateEnv", map[string]any{
		"projectId": projectId,
		"klConfig":  klFile,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	// fmt.Println(string(respData))

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

func GetLoadMaps() (map[string]string, CSResp, CSResp, error) {

	kt, err := client.GetKlFile(nil)
	if err != nil {
		return nil, nil, nil, err
	}

	env, err := EnsureEnv(nil)

	cookie, err := getCookie()
	if err != nil {
		return nil, nil, nil, err
	}

	projectName, err := client.CurrentProjectName()
	if err != nil {
		return nil, nil, nil, err
	}

	respData, err := klFetch("cli_getConfigSecretMap", map[string]any{
		"projectName": projectName,
		"envName":     env.Name,
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
		return nil, nil, nil, err
	}

	fromResp, err := GetFromResp[EnvRsp](respData)
	if err != nil {
		return nil, nil, nil, err
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

	for _, v := range *&fromResp.Configs {
		ent := cmap[v.ConfigName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		cmap[v.ConfigName][v.Key].Value = v.Value
	}

	for _, v := range *&fromResp.Secrets {
		ent := smap[v.SecretName][v.Key]
		if ent != nil {
			result[ent.Key] = v.Value
		}

		smap[v.SecretName][v.Key].Value = v.Value
	}

	return result, cmap, smap, nil
}
