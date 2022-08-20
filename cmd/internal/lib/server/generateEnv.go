package server

import (
	// "encoding/json"
	"fmt"
)

func GenerateEnv() ([]Config, error) {
	klFile, err := GetKlFile(nil)
	if err != nil {
		return nil, err
	}

	// klJson, err := json.Marshal(klFile)

	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId, err := currentProjectId()
	if err != nil {
		return nil, err
	}

	// fmt.Println(string(klJson))

	respData, err := klFetch("cli_generateEnv", map[string]any{
		"projectId": projectId,
		"klConfig":  klFile,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	fmt.Println(string(respData))

	// type Response struct {
	// 	CoreConfigs []Config `json:"envVars"`
	// }
	// var resp Response
	// err = json.Unmarshal(respData, &resp)
	// if err != nil {
	// 	return nil, err
	// }

	// return resp.CoreConfigs, nil
	return nil, nil
}
