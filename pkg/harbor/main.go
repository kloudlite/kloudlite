package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	t "kloudlite.io/pkg/types"
)

type Config interface {
	GetHarborConfig() (username string, password string, registryUrl string)
}

type harbor struct {
	username    string
	password    string
	registryUrl *url.URL
}

func (h *harbor) withAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.username, h.password)
}

func (h *harbor) CreateUserAccount(ctx context.Context, projectName string, name string) error {
	b, err := h.checkIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if b != nil && !*b {
		return errors.Newf("project(name=%s) not found", name)
	}

	// create account
	s, err := fn.CleanerNanoid(32)
	if err != nil {
		return errors.NewEf(err, "could not create nanoid")
	}
	body := t.M{
		"secret":      s,
		"name":        name,
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
		return errors.NewEf(err, "could not unmarshal request body")
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", h.registryUrl.String(), "robots"), bytes.NewBuffer(b2))
	h.withAuth(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewEf(err, "while making request")
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	return errors.Newf("could not create harbor account as received (statuscode=%d)", resp.StatusCode)
}

func (h *harbor) DeleteUserAccount(ctx context.Context, robotAccId int) error {
	if _, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s/%d", h.registryUrl.String(), "robots", robotAccId), nil); err != nil {
		return errors.NewEf(err, "could not delete harbor robot account")
	}
	return nil
}

func (h *harbor) checkIfProjectExists(ctx context.Context, name string) (*bool, error) {
	q := h.registryUrl.Query()
	q.Add("project_name", name)
	h.registryUrl.RawQuery = q.Encode()
	r, err := http.NewRequest(http.MethodHead, h.registryUrl.String(), nil)
	if err != nil {
		return nil, errors.NewEf(err, "while building http request")
	}
	r2, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, errors.NewEf(err, "while making request to check if project name already exists")
	}

	if r2.StatusCode == http.StatusOK {
		return fn.NewBool(true), nil
	}
	return fn.NewBool(false), nil
}

func (h *harbor) CreateProject(ctx context.Context, name string) error {
	b, err := h.checkIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if b != nil && *b {
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
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", h.registryUrl.String(), "projects"), bytes.NewBuffer(bbody))
	if err != nil {
		return errors.NewEf(err, "could not build request")
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.username, h.password)

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
	if b != nil && *b {
		return nil
	}
	if b != nil && !*b {
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
		registryUrl: u,
	}, nil
}
