package server

import (
	"encoding/json"
	"errors"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"net/http"
	"time"

	nanoid "github.com/matoous/go-nanoid/v2"
)

type User struct {
	UserId string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
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
			file, err := client.GetContextFile()
			if err != nil {
				return err
			}
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Cookie", loginStatusResponse.RemoteLogin.AuthHeader)
			cookie, _ := req.Cookie("hotspot-session")
			file.Session = cookie.Value
			err = client.WriteContextFile(*file)
			return err
		}
		if loginStatusResponse.RemoteLogin.Status == "failed" {
			return errors.New("remote login failed")
		}
		if loginStatusResponse.RemoteLogin.Status == "pending" {
			s := spinner.NewSpinner("waiting for login to complete")
			s.Start()
			time.Sleep(time.Second * 2)
			s.Stop()
			continue
		}
	}
}

func getCookie() (string, error) {

	file, err := client.GetContextFile()

	if err != nil {
		return "", err
	}

	if file.Session == "" {
		return "",
			errors.New("you are not logged in yet. please login using \"kl auth login\"")
	}

	return file.GetCookieString(), nil
}

type Response[T any] struct {
	Data   T       `json:"data"`
	Errors []error `json:"errors"`
}

func GetFromResp[T any](respData []byte) (*T, error) {
	var resp Response[T]
	err := json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, resp.Errors[0]
	}
	return &resp.Data, nil
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

type ItemList[T any] struct {
	Edges Edges[T] `json:"edges"`
}

func GetFromRespForEdge[T any](respData []byte) ([]T, error) {

	resp, err := GetFromResp[ItemList[T]](respData)
	if err != nil {
		return nil, err
	}

	var data []T
	for _, v := range resp.Edges {
		data = append(data, v.Node)
	}

	return data, nil
}
