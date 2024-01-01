package server

import (
	"encoding/json"
)

func GetCurrentUser() (*User, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getCurrentUser", map[string]any{}, &cookie)

	if err != nil {
		return nil, err
	}
	type Resp struct {
		User   User    `json:"data"`
		Errors []error `json:"errors"`
	}
	var resp Resp
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, resp.Errors[0]
	}
	return &resp.User, nil
}
