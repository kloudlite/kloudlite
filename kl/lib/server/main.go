package server

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/kloudlite/kl/lib/common"
	nanoid "github.com/matoous/go-nanoid/v2"
)

type User struct {
	UserId string `json:"userId"`
}

type Account struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Id          string `json:"id"`
	ReadableId  string `json:"readableId"`
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

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

var authSecret string

func CreateRemoteLogin() (loginId string, err error) {
	authSecret, err = nanoid.New(32)
	if err != nil {
		return "", err
	}

	respData, err := klFetch("cli_createRemoteLogin", map[string]any{
		"secret": authSecret,
	}, nil)

	if err != nil {
		return "", err
	}

	type Response struct {
		Id string `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

func Login(loginId string) error {
	for {
		respData, err := klFetch("cli_getRemoteLogin", map[string]any{
			"loginId": loginId,
			"secret":  authSecret,
		}, nil)

		if err != nil {
			return err
		}
		type Response struct {
			RemoteLogin struct {
				Status     string `json:"status"`
				AuthHeader string `json:"authHeader"`
			} `json:"data"`
		}
		var loginStatusResponse Response
		err = json.Unmarshal(respData, &loginStatusResponse)
		if err != nil {
			return err
		}
		if loginStatusResponse.RemoteLogin.Status == "succeeded" {

			file, err := GetContextFile()
			if err != nil {
				return err
			}

			file.Session = loginStatusResponse.RemoteLogin.AuthHeader

			err = WriteContextFile(*file)
			return err
		}
		if loginStatusResponse.RemoteLogin.Status == "failed" {
			return errors.New("remote login failed")
		}
		if loginStatusResponse.RemoteLogin.Status == "pending" {
			time.Sleep(time.Second * 2)
			continue
		}
	}
}

func CurrentAccountId() (string, error) {

	file, err := GetContextFile()

	if err != nil {
		return "", err
	}

	if file.AccountId == "" {
		return "",
			errors.New("no accounts is selected yet. please select one using \"kl use account\"")
	}

	return file.AccountId, nil
}

func CurrentDeviceId() (string, error) {

	file, err := GetContextFile()

	if err != nil {
		return "", err
	}

	if file.DeviceId == "" {
		return "",
			errors.New("no device is selected yet. please select one using \"kl use device\"")
	}

	return file.DeviceId, nil
}

func CurrentProjectId() (string, error) {

	file, err := GetContextFile()

	if err != nil {
		return "", err
	}

	if file.ProjectId == "" {
		return "",
			errors.New("no project is selected yet. please select one using \"kl use project\"")
	}

	return file.ProjectId, nil
}

func getCookie() (string, error) {

	file, err := GetContextFile()

	if err != nil {
		return "", err
	}

	if file.Session == "" {
		return "",
			errors.New("you are not logged in yet. please login using \"kl auth login\"")
	}

	return file.Session, nil
}

func GetAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getAccountMemeberships", map[string]any{
		"secret": authSecret,
	}, &cookie)

	if err != nil {
		return nil, err
	}
	type Response struct {
		AuthMe struct {
			AccountMemberships []struct {
				Account Account `json:"account"`
			} `json:"accountMemberships"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	accounts := make([]Account, len(resp.AuthMe.AccountMemberships))
	for i, v := range resp.AuthMe.AccountMemberships {
		accounts[i] = v.Account
	}
	return accounts, nil
}

func GetProjects(options ...common.Option) ([]Project, error) {
	accountId := common.GetOption(options, "accountId")

	cookie, err := getCookie()

	if err != nil {
		return nil, err
	}

	if accountId == "" {
		accountId, err = CurrentAccountId()

		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_getProjects", map[string]any{
		"accountId": accountId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		FinanceAccount struct {
			Projects []Project `json:"projects"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return resp.FinanceAccount.Projects, nil
}

func GetApps(options ...common.Option) ([]App, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId := common.GetOption(options, "projectId")
	if projectId == "" {
		projectId, err = CurrentProjectId()
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

func GetEnvs(appId string) (string, error) {
	cookie, err := getCookie()
	if err != nil {
		return "", err
	}

	respData, err := klFetch("cli_getEnv", map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return "", err
	}

	type Response struct {
		Envs string `json:"data"`
	}

	var resp Response

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", err
	}

	return resp.Envs, nil
}
