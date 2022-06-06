package server

import (
	"encoding/json"
	"errors"
	"fmt"
	nanoid "github.com/matoous/go-nanoid/v2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type User struct {
	UserId string `json:"userId"`
}

func Me() (*User, error) {
	return &User{}, nil
}

var authSecret string

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
	for true {
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
			var dirName string
			dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
			if !ok {
				dirName, err = os.UserHomeDir()
				if err != nil {
					return err
				}
			}
			configFolder := fmt.Sprintf("%s/.kl", dirName)
			if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(configFolder, os.ModePerm)
				if err != nil {
					log.Println(err)
				}
			}
			err := ioutil.WriteFile(fmt.Sprintf("%v/session", configFolder), []byte(loginStatusResponse.Data.RemoteLogin.AuthHeader), 0644)
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

	return nil
}
func SelectProject(projectId string) error {
	return nil
}
func GetProjects() {

}
func LoadApp() {

}
func GetApps() {

}
