package server

import (
	"encoding/json"
	"fmt"
	fn "github.com/kloudlite/kl/pkg/functions"
	"strings"
)

type App struct {
	//IsLambda   bool   `json:"isLambda"`
	//Id         string `json:"id"`
	//Name       string `json:"name"`
	//ReadableId string `json:"readableId"`
	//Containers []struct {
	//	Name    string `json:"name"`
	//	EnvVars []struct {
	//		Key   string `json:"key"`
	//		Value struct {
	//			Key   string `json:"key"`
	//			Ref   string `json:"ref"`
	//			Type  string `json:"type"`
	//			Value string `json:"value"`
	//		} `json:"value"`
	//	} `json:"envVars"`
	//} `json:"containers"`
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
}

func ListApps(options ...fn.Option) ([]App, error) {

	var err error
	projectName, err := EnsureProject(options...)
	if err != nil {
		return nil, err
	}

	envName, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	fmt.Println(envName.Name, projectName, envName.IsEnvironment)

	respData, err := klFetch("cli_listApps", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"project": map[string]any{
			"type":  "name",
			"value": strings.TrimSpace(projectName),
		},
		"scope": map[string]any{
			"type":  "environmentName",
			"value": strings.TrimSpace(envName.Name),
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[App](respData); err != nil {
		return nil, err
	} else {
		fmt.Println(fromResp)
		return fromResp, nil
	}
}

func GetApp(appId string) (*App, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getApp", map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreApp App `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.CoreApp, nil
}
