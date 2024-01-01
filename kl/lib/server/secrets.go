package server

import (
	"encoding/json"

	common_util "github.com/kloudlite/kl/lib/common"
)

type Secret struct {
	Entries []CSEntry `json:"entries"`
	Name    string    `json:"name"`
	Id      string    `json:"id"`
}

func GetSecrets(options ...common_util.Option) ([]Secret, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId := common_util.GetOption(options, "projectId")
	if projectId == "" {
		projectId, err = CurrentProjectId()
		if err != nil {
			return nil, err
		}
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

func GetSecret(id string) (*Config, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getSecret", map[string]any{
		"secretId": id,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		CoreSecret Config `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.CoreSecret, nil
}
