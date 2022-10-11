package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"operators.kloudlite.io/lib/errors"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

type harborRobotUser struct {
	UpdateTime   time.Time `json:"update_time"`
	Description  string    `json:"description"`
	Level        string    `json:"level"`
	Editable     bool      `json:"editable"`
	CreationTime time.Time `json:"creation_time"`
	ExpiresAt    int       `json:"expires_at"`
	Name         string    `json:"name"`
	Secret       string    `json:"secret"`
	Disable      bool      `json:"disable"`
	Duration     int       `json:"duration"`
	Id           int       `json:"id"`
	Permissions  []struct {
		Access []struct {
			Action   string `json:"action"`
			Resource string `json:"resource"`
			Effect   string `json:"effect"`
		} `json:"access"`
		Kind      string `json:"kind"`
		Namespace string `json:"namespace"`
	} `json:"permissions"`
}

var dockerRepoMinACLs = []map[string]any{
	{
		"action":   "push",
		"resource": "repository",
	},
	{
		"action":   "pull",
		"resource": "repository",
	},
}

func (h *Client) CreateUserAccount(ctx context.Context, projectName, userName, password string) (*User, error) {
	body := map[string]any{
		"secret":      password,
		"name":        userName,
		"level":       "project",
		"duration":    -1,
		"description": "created by kloudlite operator",
		"permissions": []map[string]any{
			{
				"access":    dockerRepoMinACLs,
				"kind":      "project",
				"namespace": projectName,
			},
		},
	}

	b2, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	fmt.Println("body: ", string(b2))

	req, err := h.NewAuthzRequest(ctx, http.MethodPost, "/robots", bytes.NewBuffer(b2))
	if err != nil {
		return nil, errors.NewEf(err, "building requests for creating robot account")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, errors.Newf("bad status code (%d), with error message %s", resp.StatusCode, rbody)
	}

	var hUser harborRobotUser
	if err := json.Unmarshal(rbody, &hUser); err != nil {
		return nil, errors.NewEf(err, "could not unmarshal into harborRobotUser")
	}

	req, err = h.NewAuthzRequest(
		ctx, http.MethodPatch, fmt.Sprintf("/robots/%d", hUser.Id),
		bytes.NewBuffer([]byte(fmt.Sprintf("{\"secret\":\"%s\"}", password))),
	)
	if err != nil {
		return nil, err
	}

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp2.StatusCode == http.StatusOK {
		return &User{
			Id:       hUser.Id,
			Name:     hUser.Name,
			Location: resp.Header.Get("Location"),
		}, nil
	}
	return nil, errors.New("bad status code")
}

func (h *Client) UpdateUserAccount(ctx context.Context, user *User, enabled bool) error {
	// ASSERT: artifacts-harbor update is terrible, they required an entire object, instead of only diffs
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, user.Location, nil)
	if err != nil {
		return errors.NewEf(err, "building requests for updating robot account")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r harborRobotUser
	if err := json.Unmarshal(respBody, &r); err != nil {
		return err
	}

	r.Disable = !enabled

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err = h.NewAuthzRequest(ctx, http.MethodPut, user.Location, bytes.NewBuffer(b))
	if err != nil {
		return errors.NewEf(err, "building requests for updating robot account")
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("could not update user account as received statuscode=%d because %s", resp.StatusCode, respBody)
	}

	return nil
}

func (h *Client) DeleteUserAccount(ctx context.Context, user *User) error {
	req, err := h.NewAuthzRequest(ctx, http.MethodDelete, user.Location, nil)
	if err != nil {
		return errors.NewEf(err, "making request to delete harbor account")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewEf(err, "could not delete Client robot account")
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		// ASSERT: silent exit, as artifacts-harbor user account is already deleted
		return nil
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return errors.Newf("bad status code (%d), with error message, %s", resp.StatusCode, msg)
}

func (h *Client) CheckIfUserAccountExists(ctx context.Context, user *User) (bool, error) {
	if user == nil || user.Location == "" {
		return false, nil
	}
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, user.Location, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict, nil
}
