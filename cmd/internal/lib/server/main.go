package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"kloudlite.io/cmd/internal/common"
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

func getConfigFolder() (configFolder string, err error) {
	var dirName string
	dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		dirName, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	configFolder = fmt.Sprintf("%s/.kl", dirName)
	if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(configFolder, os.ModePerm)
		if err != nil {
			common.PrintError(err)
		}
	}
	return configFolder, nil
}

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
			configFolder, err := getConfigFolder()
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(fmt.Sprintf("%v/session", configFolder), []byte(loginStatusResponse.RemoteLogin.AuthHeader), 0644)
			if err != nil {
				return err
			}
			return nil
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
	// klfile, err := GetKlFile(nil)
	// if err != nil {
	// 	return "", err
	// }

	// if klfile.AccountId == "" {
	// 	return "", errors.New("no project selected")
	// }

	// return klfile.AccountId, nil
	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}
	var file []byte
	file, err = ioutil.ReadFile(fmt.Sprintf("%s/account", folder))
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func CurrentProjectId() (string, error) {
	// klfile, err := GetKlFile(nil)
	// if err != nil {
	// 	return "", err
	// }

	// if klfile.ProjectId == "" {
	// 	return "", errors.New("no project selected")
	// }

	// return klfile.ProjectId, nil

	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}

	var file []byte

	file, err = ioutil.ReadFile(fmt.Sprintf("%s/project", folder))
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func getCookie() (string, error) {
	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}
	var file []byte

	file, err = ioutil.ReadFile(fmt.Sprintf("%s/session", folder))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(file)), nil
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
