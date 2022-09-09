package server

import (
	// "encoding/json"
	"encoding/json"
)

type GeneratedEnvs struct {
	EnvVars    map[string]string `json:"envVars"`
	MountFiles map[string]string `json:"mountFiles"`
}

func GenerateEnv() (*GeneratedEnvs, error) {
	klFile, err := GetKlFile(nil)
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

	projectId, err := CurrentProjectId()
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
