package server

import "encoding/json"

type Secret struct {
	Description string
	Entries     []CSEntry
	Id          string
	Name        string
	Namsespace  string
	status      string
}

func GetSecrets() ([]Secret, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId, err := currentProjectId()
	if err != nil {
		return nil, err
	}

	respData, err := gql(`
	query Core_secrets($projectId: ID!) {
		core_secrets(projectId: $projectId) {
			entries {
				key
			}
			id
			name
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
			CoreSecrets []Secret `json:"core_secrets"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data.CoreSecrets, nil
}
