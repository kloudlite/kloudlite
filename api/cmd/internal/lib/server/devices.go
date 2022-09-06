package server

import (
	"encoding/json"
	"errors"

	"kloudlite.io/cmd/internal/common"
)

type DAccount struct {
	Devices []Device `json:"devices"`
}

type DApp struct {
	ReadableId string `json:"readableId"`
	Name       string `json:"name"`
	Id         string `json:"id"`
}

type Device struct {
	Region string `json:"region"`
	Ports  []struct {
		Port       int `json:"port"`
		TargetPort int `json:"targetPort"`
	} `json:"ports"`
	Name          string            `json:"name"`
	Id            string            `json:"id"`
	Intercepted   []DApp            `json:"interceptingServices"`
}

func GetDevices(options ...common.Option) ([]Device, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	accountId := common.GetOption(options, "accountId")
	if accountId == "" {
		accountId, err = CurrentAccountId()

		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_getDevices", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Account map[string]DAccount `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	account := resp.Account[accountId]

	return account.Devices, nil
}

func InterceptApp(devieId, appId string) error {
	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_interceptApp", map[string]any{
		"deviceId": devieId,
		"appId":    appId,
	}, &cookie)

	if err != nil {
		return err
	}

	var resp struct {
		Inercepted bool `json:"data"`
	}

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return err
	}

	if resp.Inercepted {
		return nil
	}

	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
}

func CloseInterceptApp(appId string) error {
	cookie, err := getCookie()
	if err != nil {
		return err
	}

	respData, err := klFetch("cli_closeIntercept", map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return err
	}

	var resp struct {
		Inercepted bool `json:"data"`
	}

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return err
	}

	if resp.Inercepted {
		return nil
	}

	return errors.New("SOMETHING WENT WRONG... PLEASE TRY AGAIN")
}
