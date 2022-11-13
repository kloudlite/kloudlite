package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"operators.kloudlite.io/lib/errors"
	hTypes "operators.kloudlite.io/lib/harbor/internal/types"
)

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Password string `json:"-"`
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

func (h *Client) changePassword(ctx context.Context, user *User) (*User, error) {
	return nil, nil
}

func (h *Client) CreateUserAccount(ctx context.Context, projectName, userName string) (*User, error) {
	body := map[string]any{
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

	var hUser hTypes.RobotUser
	if err := json.Unmarshal(rbody, &hUser); err != nil {
		return nil, errors.NewEf(err, "could not unmarshal into harborRobotUser")
	}

	return &User{
		Id:       hUser.Id,
		Name:     hUser.Name,
		Password: hUser.Secret,
	}, nil

	// req, err = h.NewAuthzRequest(
	// 	ctx, http.MethodPatch, fmt.Sprintf("/robots/%d", hUser.Id),
	// 	bytes.NewBuffer([]byte(fmt.Sprintf("{\"secret\":\"%s\"}", password))),
	// )
	// if err != nil {
	// 	return nil, err
	// }
	//
	// resp2, err := http.DefaultClient.Do(req)
	// if err != nil {
	// 	return nil, err
	// }
	// if resp2.StatusCode == http.StatusOK {
	// 	return &User{
	// 		Id:       hUser.Id,
	// 		Name:     hUser.Name,
	// 		Location: resp.Header.Get("Location"),
	// 	}, nil
	// }
	// return nil, errors.New("bad status code")
}

func (h *Client) UpdateUserAccount(ctx context.Context, robotId int64, enabled bool) error {
	// ASSERT: artifacts-harbor update is terrible, they required an entire object, instead of only diffs
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/robots/%d", robotId), nil)
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

	var r hTypes.RobotUser
	if err := json.Unmarshal(respBody, &r); err != nil {
		return err
	}

	r.Disable = !enabled

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err = h.NewAuthzRequest(ctx, http.MethodPut, fmt.Sprintf("/robots/%d", robotId), bytes.NewBuffer(b))
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

func (h *Client) DeleteUserAccount(ctx context.Context, userId int64) error {
	req, err := h.NewAuthzRequest(ctx, http.MethodDelete, fmt.Sprintf("/robots/%d", userId), nil)
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

func (h *Client) CheckIfUserAccountExists(ctx context.Context, userId int64) (bool, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/robots/%d", userId), nil)
	if err != nil {
		return false, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict, nil
}

func (h *Client) FindUserAccountByName(ctx context.Context, projectName string, username string) (*User, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/projects/%s/robots", projectName), nil)
	if err != nil {
		return nil, errors.NewEf(err, "creating request")
	}

	qp := req.URL.Query()
	qp.Add("q", fmt.Sprintf("name=~%s", username))
	req.URL.RawQuery = qp.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.NewEf(err, "while calling")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var users []hTypes.RobotUser
	if err := json.Unmarshal(b, &users); err != nil {
		return nil, err
	}

	for i := range users {
		if strings.HasSuffix(users[i].Name, username) {
			return &User{
				Id:   users[i].Id,
				Name: users[i].Name,
			}, nil
		}
	}

	return nil, errors.NewHttpError(http.StatusNotFound, fmt.Sprintf("no robot user account named (%s) found", username))
}
