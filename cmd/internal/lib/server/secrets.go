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

	respData, err := klFetch("cli_getSecrets", map[string]any{
		"projectId": projectId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreSecrets []Secret `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return resp.CoreSecrets, nil
}
