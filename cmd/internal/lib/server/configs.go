package server

import (
	"encoding/json"
)

type CSEntry struct {
	Value string
	Key   string
}

type Config struct {
	Description string
	Entries     []CSEntry
	Id          string
	Name        string
	Namsespace  string
	status      string
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

	respData, err := gql(`
	query Core_configs($projectId: ID!) {
		core_configs(projectId: $projectId) {
			entries {
				value
				key
			}
			description
			id
			name
			namespace
			status
		}
	}
	`, map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		Data struct {
			CoreConfigs []Config `json:"core_configs"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data.CoreConfigs, nil
}
