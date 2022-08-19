package server

import (
	"encoding/json"
)

type CSEntry struct {
	Value string
	Key   string
}

type Config struct {
	Entries []CSEntry
	Id      string
	Name    string
}

type ConfigORSecret struct {
	Entries []CSEntry
	Id      string
	Name    string
}

func GetConfigs() ([]Config, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId, err := currentProjectId()
	if err != nil {
		return nil, err
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
