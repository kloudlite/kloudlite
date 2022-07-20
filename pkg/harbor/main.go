package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kloudlite.io/pkg/errors"
	"net/http"
	"net/url"

	fn "kloudlite.io/pkg/functions"
	t "kloudlite.io/pkg/types"
)

type Config interface {
	GetHarborConfig() (username string, password string, registryUrl string)
}

type harbor struct {
	username    string
	password    string
	registryUrl url.URL
}

func (h *harbor) withAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.username, h.password)
}

type User struct {
	Id     int
	Name   string
	Secret string
}

func (h *harbor) CreateUserAccount(ctx context.Context, projectName string) (*User, error) {
	b, err := h.checkIfProjectExists(ctx, projectName)
	if err != nil {
		return nil, err
	}
	if !b { // i.e. project does not exist
		return nil, errors.Newf("project %s does not exist", projectName)
	}

	// create account
	s, err := fn.CleanerNanoid(32)
	if err != nil {
		return nil, err
	}
	body := t.M{
		"secret":      s,
		"name":        projectName,
		"level":       "system",
		"duration":    0,
		"description": "created by kloudlite/ci",
		"permissions": []t.M{
			{
				"access": []t.M{
					{
						"action":   "push",
						"resource": "repository",
					},
					{
						"action":   "pull",
						"resource": "repository",
					},
					{
						"action":   "delete",
						"resource": "artifact",
					},
					{
						"action":   "create",
						"resource": "helm-chart-version",
					},
					{
						"action":   "delete",
						"resource": "helm-chart-version",
					},
					{
						"action":   "create",
						"resource": "tag",
					},
					{
						"action":   "delete",
						"resource": "tag",
					},
					{
						"action":   "create",
						"resource": "artifact-label",
					},
					{
						"action":   "list",
						"resource": "artifact",
					},
					{
						"action":   "list",
						"resource": "repository",
					},
				},
				"kind":      "project",
				"namespace": projectName,
			},
		},
	}

	b2, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", h.registryUrl.String(), "robots"),
		bytes.NewBuffer(b2),
	)
	h.withAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var user User
	if err := json.Unmarshal(rbody, &user); err != nil {
		return nil, errors.NewEf(err, "could not unmarshal into harborUser")
	}

	if resp.StatusCode == http.StatusCreated {
		return &user, nil
	}
	return nil, errors.Newf("could not create user account as received statuscode=%d", resp.StatusCode)
}

func (h *harbor) DeleteUserAccount(ctx context.Context, robotAccId int) error {
	if _, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s/%d", h.registryUrl.String(), "robots", robotAccId),
		nil,
	); err != nil {
		return errors.NewEf(err, "could not delete harbor robot account")
	}
	return nil
}

func (h *harbor) checkIfProjectExists(ctx context.Context, name string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, fmt.Sprintf("%s/projects", h.registryUrl.String()), nil)
	if err != nil {
		return false, errors.NewEf(err, "while building http request")
	}
	q := req.URL.Query()
	q.Add("project_name", name)
	req.URL.RawQuery = q.Encode()
	h.withAuth(req)
	fmt.Println("checkprojects: url=>", req.URL.String())
	r2, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.NewEf(err, "while making request to check if project name already exists")
	}

	if r2.StatusCode == http.StatusOK {
		return true, nil
	}
	all, err := io.ReadAll(r2.Body)
	fmt.Println("all:", string(all), "err:", err, "statuscode:", r2.StatusCode)
	return false, nil
}

func (h *harbor) CreateProject(ctx context.Context, name string) error {
	b, err := h.checkIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if b { // ie. project exists
		return nil
	}

	body := t.M{
		"project_name": name,
		"public":       false,
	}
	bbody, err := json.Marshal(body)
	if err != nil {
		return errors.NewEf(err, "could not unmarshal req body")
	}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", h.registryUrl.String(), "projects"),
		bytes.NewBuffer(bbody),
	)
	if err != nil {
		return errors.NewEf(err, "could not build request")
	}
	fmt.Println("url:", req.URL)
	h.withAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return errors.NewEf(err, "while making request")
	}
	if resp.StatusCode == http.StatusCreated {
		return nil
	}
	return errors.Newf("could not create harbor project as received (statuscode=%d)", resp.StatusCode)
}

func (h *harbor) DeleteProject(ctx context.Context, name string) error {
	b, err := h.checkIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if !b {
		return errors.Newf("harbor project(name=%s) does not exist", name)
	}

	u, err := h.registryUrl.Parse(name)
	if err != nil {
		return errors.NewEf(err, "could not join url path param")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u.String(), nil)
	if err != nil {
		return errors.NewEf(err, "while building http request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewEf(err, "while making request")
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.Newf("could not delete harbor project as received (statuscode=%d)", resp.StatusCode)
}

func NewClient(cfg Config) (Harbor, error) {
	username, password, registryUrl := cfg.GetHarborConfig()
	u, err := url.Parse(registryUrl)
	if err != nil {
		return nil, errors.NewEf(err, "registryUrl is not a valid url")
	}
	return &harbor{
		username:    username,
		password:    password,
		registryUrl: *u,
	}, nil
}
