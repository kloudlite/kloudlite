package apiclient

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	nanoid "github.com/matoous/go-nanoid/v2"
)

func (apic *apiClient) GetCurrentUser() (*User, error) {
	cookie, err := getCookie()
	if err != nil && cookie == "" {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_getCurrentUser", map[string]any{}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}
	type Resp struct {
		User   User    `json:"data"`
		Errors []error `json:"errors"`
	}
	var resp Resp
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, functions.NewE(err)
	}
	if len(resp.Errors) > 0 {
		return nil, resp.Errors[0]
	}
	return &resp.User, nil
}

func (apic *apiClient) CreateRemoteLogin() (loginId string, err error) {
	authSecret, err = nanoid.New(32)
	if err != nil {
		return "", functions.NewE(err)
	}

	respData, err := klFetch("cli_createRemoteLogin", map[string]any{
		"secret": authSecret,
	}, nil)

	if err != nil {
		return "", functions.NewE(err)
	}

	type Response struct {
		Id string `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return "", functions.NewE(err)
	}
	return resp.Id, nil
}

func (apic *apiClient) Login(loginId string) error {
	for {
		respData, err := klFetch("cli_getRemoteLogin", map[string]any{
			"loginId": loginId,
			"secret":  authSecret,
		}, nil)

		if err != nil {
			return functions.NewE(err)
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
			return functions.NewE(err)
		}
		if loginStatusResponse.RemoteLogin.Status == "succeeded" {
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("Cookie", loginStatusResponse.RemoteLogin.AuthHeader)
			cookie, _ := req.Cookie("hotspot-session")

			return fileclient.SaveAuthSession(cookie.Value)
		}
		if loginStatusResponse.RemoteLogin.Status == "failed" {
			return functions.Error("remote login failed")
		}
		if loginStatusResponse.RemoteLogin.Status == "pending" {
			spinner.Client.UpdateMessage("waiting for login to complete")
			time.Sleep(time.Second * 2)
			spinner.Client.Pause()
			continue
		}
	}
}
