package server

import (
	"encoding/json"
	"github.com/kloudlite/kl/domain/client"
	common_util "github.com/kloudlite/kl/pkg/functions"
)

type App struct {
	IsLambda   bool   `json:"isLambda"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	ReadableId string `json:"readableId"`
	Containers []struct {
		Name    string `json:"name"`
		EnvVars []struct {
			Key   string `json:"key"`
			Value struct {
				Key   string `json:"key"`
				Ref   string `json:"ref"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"value"`
		} `json:"envVars"`
	} `json:"containers"`
}

func GetApps(options ...common_util.Option) ([]App, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId := common_util.GetOption(options, "projectId")
	if projectId == "" {
		projectId, err = client.CurrentProjectName()
		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_getApps", map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreApps []App `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return resp.CoreApps, nil
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
