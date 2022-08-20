package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Secret   string `json:"secret"`
	Location string `json:"location"`
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

func (h *Client) CreateUserAccount(ctx context.Context, projectName string, userName string) (*User, error) {
	// create account
	s := fn.CleanerNanoid(48)

	body := map[string]any{
		"secret":      s,
		"name":        userName,
		"level":       "project",
		"duration":    0,
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

	if resp.StatusCode == http.StatusCreated {
		var user User
		if err := json.Unmarshal(rbody, &user); err != nil {
			return nil, errors.NewEf(err, "could not unmarshal into harborUser")
		}
		user.Location = resp.Header.Get("Location")
		return &user, nil
	}

	return nil, errors.Newf("bad status code (%d), with error message %s", resp.StatusCode, rbody)
}

type harborRobotUserTT struct {
	Description string `json:"description"`
	Disable     bool   `json:"disable"`
	Duration    int    `json:"duration"`
	ExpiresAt   int    `json:"expires_at"`
	Id          int    `json:"id"`
	Level       string `json:"level"`
	Name        string `json:"name"`
	Permissions []struct {
		Access []struct {
			Action   string `json:"action"`
			Resource string `json:"resource"`
		} `json:"access"`
		Kind      string `json:"kind"`
		Namespace string `json:"namespace"`
	} `json:"permissions"`
	UpdateTime time.Time `json:"update_time"`
}

func (h *Client) UpdateUserAccount(ctx context.Context, user *User, enabled bool) error {
	// ASSERT: harbor update is super super bad, they required an entire object, instead of only diffs
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

	var r harborRobotUserTT
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

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.Newf("could not update user account as received statuscode=%d because %s", resp.StatusCode, respBody)
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
		// ASSERT: silent exit, as harbor user account is already deleted
		return nil
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return errors.Newf("bad status code (%d), with error message, %s", resp.StatusCode, msg)
}

func (h *Client) CheckIfUserAccountExists(ctx context.Context, user *User) (bool, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, user.Location, nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}
