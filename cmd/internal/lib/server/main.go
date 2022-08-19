package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
	"kloudlite.io/cmd/internal/constants"
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
			log.Println(err)
		}
	}
	return configFolder, nil
}

func CreateRemoteLogin() (loginId string, err error) {
	authSecret, err = nanoid.New(32)
	if err != nil {
		return "", err
	}

	respData, err := gql(`
		mutation Auth_createRemoteLogin($secret: String) {
			auth_createRemoteLogin(secret: $secret)
		}
		`, map[string]any{
		"secret": authSecret,
	}, nil)

	if err != nil {
		return "", err
	}

	type Response struct {
		Data struct {
			Id string `json:"auth_createRemoteLogin"`
		} `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", err
	}
	return resp.Data.Id, nil
}

func Login(loginId string) error {
	for {
		respData, err := gql(`
		query Auth_getRemoteLogin($loginId: String!, $secret: String!) {
  			auth_getRemoteLogin(loginId: $loginId, secret: $secret) {
    			status
				authHeader
  			}
		}
		`, map[string]any{
			"loginId": loginId,
			"secret":  authSecret,
		}, nil)
		if err != nil {
			return err
		}
		type Response struct {
			Data struct {
				RemoteLogin struct {
					Status     string `json:"status"`
					AuthHeader string `json:"authHeader"`
				} `json:"auth_getRemoteLogin"`
			} `json:"data"`
		}
		var loginStatusResponse Response
		err = json.Unmarshal(respData, &loginStatusResponse)
		if err != nil {
			return err
		}
		if loginStatusResponse.Data.RemoteLogin.Status == "succeeded" {
			configFolder, err := getConfigFolder()
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(fmt.Sprintf("%v/session", configFolder), []byte(loginStatusResponse.Data.RemoteLogin.AuthHeader), 0644)
			if err != nil {
				return err
			}
			return nil
		}
		if loginStatusResponse.Data.RemoteLogin.Status == "failed" {
			return errors.New("remote login failed")
		}
		if loginStatusResponse.Data.RemoteLogin.Status == "pending" {
			time.Sleep(time.Second * 2)
			continue
		}
	}
}

func currentAccountId() (string, error) {
	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}
	var file []byte
	count := 0
	for {
		if count > 2 {
			return "", err
		}
		file, err = ioutil.ReadFile(fmt.Sprintf("%s/account", folder))
		if err == nil {
			break
		}
		exec.Command(constants.CMD_NAME, "accounts").Run()
		count++
	}

	return string(file), nil
}

func currentProjectId() (string, error) {
	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}

	var file []byte
	count := 0
	for {
		if count > 2 {
			return "", err
		}
		file, err = ioutil.ReadFile(fmt.Sprintf("%s/project", folder))
		if err == nil {
			break
		}
		exec.Command(constants.CMD_NAME, "project", "list").Run()
		count++
	}

	return string(file), nil
}

func getCookie() (string, error) {
	folder, err := getConfigFolder()
	if err != nil {
		return "", err
	}
	var file []byte

	count := 0
	for {
		if count > 2 {
			return "", err
		}
		file, err = ioutil.ReadFile(fmt.Sprintf("%s/session", folder))
		if err == nil {
			break
		}
		exec.Command(constants.CMD_NAME, "login").Run()
		count++
	}

	return strings.TrimSpace(string(file)), nil
}

func GetAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := gql(`
		query AccountMemberships {
          auth_me {
            accountMemberships {
              account {
                id
                name
              }
            }
          }
        }
		`, map[string]any{
		"secret": authSecret,
	}, &cookie)

	if err != nil {
		return nil, err
	}
	type Response struct {
		Data struct {
			AuthMe struct {
				AccountMemberships []struct {
					Account Account `json:"account"`
				} `json:"accountMemberships"`
			} `json:"auth_me"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	accounts := make([]Account, len(resp.Data.AuthMe.AccountMemberships))
	for i, v := range resp.Data.AuthMe.AccountMemberships {
		accounts[i] = v.Account
	}
	return accounts, nil
}

func GetProjects() ([]Project, error) {
	cookie, err := getCookie()

	if err != nil {
		return nil, err
	}

	accountId, err := currentAccountId()

	if err != nil {

		return nil, err
	}

	respData, err := gql(`
		query Projects($accountId: ID!) {
          finance_account(accountId: $accountId) {
            projects {
              id
              readableId
			  displayName
			  name
            }
          }
        }
		`, map[string]any{
		"accountId": accountId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		Data struct {
			FinanceAccount struct {
				Projects []Project `json:"projects"`
			} `json:"finance_account"`
		} `json:"data"`
	}
	var resp Response
	fmt.Println(resp)
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data.FinanceAccount.Projects, nil
}

func GetApps() ([]App, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	projectId, err := currentProjectId()

	if err != nil {
		return nil, err
	}

	// count := 0
	// for {
	// 	if count > 2 {
	// 		return nil, err
	// 	}
	// 	if err == nil {
	// 		break
	// 	}
	// 	exec.Command(constants.CMD_NAME, "projects").Run()
	// 	count++
	// }

	respData, err := gql(`
		query Core_apps($projectId: ID!) {
          core_apps(projectId: $projectId) {
            id
            name
            readableId
            containers {
              name
              
              envVars {
                key
                value {
                  ref
                  key
                  type
                }
              }
            }
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
			CoreApps []App `json:"core_apps"`
		} `json:"data"`
	}
	var resp Response
	// fmt.Println(resp)
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data.CoreApps, nil
}

func GetApp(appId string) (*App, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}
	respData, err := gql(`
		query Core_app($appId: ID!) {
          core_app(appId: $appId) {
            id
            name
            readableId
            containers {
              name
              
              envVars {
                key
                value {
                  ref
                  key
                  type
                  value
                }
              }
            }
          }
        }
		`, map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return nil, err
	}

	type Response struct {
		Data struct {
			CoreApp App `json:"core_app"`
		} `json:"data"`
	}
	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Data.CoreApp, nil
}

func GetEnvs(appId string) (string, error) {
	cookie, err := getCookie()
	if err != nil {
		return "", err
	}

	respData, err := gql(`
		query Core_app($appId: ID!) {
			core_app_envs(appId: $appId) 
		}	
		`, map[string]any{
		"appId": appId,
	}, &cookie)

	if err != nil {
		return "", err
	}

	type Response struct {
		Data struct {
			Envs string `json:"core_app_envs"`
		} `json:"data"`
	}

	var resp Response

	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", err
	}

	return resp.Data.Envs, nil
}

func LoadApp() {
}
