package server

import "errors"

type Account struct {
	Metadata struct {
		Name string `json:"name"`
	}
	DisplayName string `json:"displayName"`
}

func CurrentAccountName() (string, error) {
	file, err := GetContextFile()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New("noSelectedAccount")
	}
	return file.AccountName, nil
}

func GetAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}
	type AccList []Account
	if fromResp, err := GetFromResp[AccList](respData); err != nil {
		return nil, err
	} else {
		return *fromResp, nil
	}
}
