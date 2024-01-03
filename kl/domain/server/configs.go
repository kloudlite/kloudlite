package server

import (
	"encoding/json"
	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type CSEntry struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

type Config struct {
	Entries []CSEntry `json:"entries"`
	Name    string    `json:"name"`
	Id      string    `json:"id"`
}

type ConfigORSecret struct {
	Entries []CSEntry `json:"entries"`
	Name    string    `json:"name"`
}

func GetConfigs(options ...fn.Option) ([]Config, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId := fn.GetOption(options, "projectId")
	if projectId == "" {
		projectId, err = client.CurrentProjectName()
		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_getConfigs", map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreConfigs []Config `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return resp.CoreConfigs, nil
}

func GetConfig(id string) (*Config, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getConfig", map[string]any{
		"configId": id,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreConfig Config `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.CoreConfig, nil
}
