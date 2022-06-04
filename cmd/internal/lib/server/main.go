package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	nanoid "github.com/matoous/go-nanoid/v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	authUrl = "https://auth.kl.local.madhouselabs.io/api"
)

type User struct {
	UserId string `json:"userId"`
}

func Me() (*User, error) {
	return &User{}, nil
}

var authSecret string

func CreateRemoteLogin() (string, error) {
	authSecret, err := nanoid.New(32)
	if err != nil {
		return "", err
	}
	body := map[string]any{
		"method": "createRemoteLogin",
		"args": []map[string]string{
			{
				"secret": authSecret,
			},
		},
	}
	marshal, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	post, err := http.Post(fmt.Sprintf("%s/login", authUrl), "application/json", bytes.NewReader(marshal))
	if err != nil {
		return "", err
	}
	type RemoteLogin struct {
		Id string `json:"id"`
	}
	var remoteLogin RemoteLogin
	err = json.NewDecoder(post.Body).Decode(&remoteLogin)
	if err != nil {
		return "", err
	}
	return remoteLogin.Id, nil
}
func Login(loginId string) error {
	body := map[string]string{
		"secret": authSecret,
	}
	marshal, err := json.Marshal(body)
	if err != nil {
		return err
	}

	for true {
		post, err := http.Post(fmt.Sprintf("%s/login", authUrl), "application/json", bytes.NewReader(marshal))
		if err != nil {
			return err
		}
		type LoginStatusResponse struct {
			Status     string `json:"status"`
			AuthHeader string `json:"auth_header"`
		}
		var loginStatusResponse LoginStatusResponse
		err = json.NewDecoder(post.Body).Decode(&loginStatusResponse)
		if err != nil {
			return err
		}
		if loginStatusResponse.Status == "success" {
			var dirName string
			dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
			if !ok {
				dirName, err = os.UserHomeDir()
				if err != nil {
					return err
				}
			}
			configFolder := fmt.Sprintf("/%s/.kl", dirName)
			if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(configFolder, os.ModePerm)
				if err != nil {
					log.Println(err)
				}
			}
			err := ioutil.WriteFile(fmt.Sprintf("%v/session", configFolder), []byte(loginStatusResponse.AuthHeader), 0644)
			if err != nil {
				return err
			}
			return nil
		}
		if loginStatusResponse.Status == "failed" {
			return errors.New("remote login failed")
		}
		if loginStatusResponse.Status == "pending" {
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
