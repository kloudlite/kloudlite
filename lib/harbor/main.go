package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"operators.kloudlite.io/lib/errors"
	fn "operators.kloudlite.io/lib/functions"
)

type Config interface {
	GetHarborConfig() (username string, password string, registryUrl string)
}

type Client struct {
	args *Args
	url  *url.URL
}

func (h *Client) withAuth(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(h.args.HarborAdminUsername, h.args.HarborAdminPassword)
}

type User struct {
	Id     int
	Name   string
	Secret string
}

func (h *Client) CreateUserAccount(ctx context.Context, projectName string) (*User, error) {
	// create account
	s := fn.CleanerNanoid(48)
	body := map[string]any{
		"secret":      s,
		"name":        projectName,
		"level":       "system",
		"duration":    0,
		"description": "created by kloudlite/ci",
		"permissions": []map[string]any{
			{
				"access": []map[string]any{
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
		fmt.Sprintf("%s/%s", h.url.String(), "robots"),
		bytes.NewBuffer(b2),
	)
	if err != nil {
		return nil, errors.NewEf(err, "building requests for creating robot account")
	}

	h.withAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("Response: ", string(rbody))
	var user User
	if err := json.Unmarshal(rbody, &user); err != nil {
		return nil, errors.NewEf(err, "could not unmarshal into harborUser")
	}
	fmt.Printf("User: %+v\n", user)

	if resp.StatusCode == http.StatusCreated {
		return &user, nil
	}
	return nil, errors.Newf("could not create user account as received statuscode=%d", resp.StatusCode)
}

func (h *Client) DeleteUserAccount(ctx context.Context, robotAccId int) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s/%d", h.args.HarborRegistryHost, "robots", robotAccId),
		nil,
	)
	if err != nil {
		return errors.NewEf(err, "could not delete Client robot account")
	}
	resp, err := http.DefaultClient.Do(req)
	// REMOVE:
	fmt.Printf("resp: %+v\n", resp)
	if err != nil {
		return errors.NewEf(err, "making request to delete harbor account")
	}
	return nil
}

func (h *Client) CheckIfAccountExists(ctx context.Context, robotAccId int) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/%s/%d", h.url.String(), "robots", robotAccId),
		nil,
	)
	if err != nil {
		return false, errors.NewEf(err, "creating request to delete robotAccount")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, errors.NewEf(err, "making request to delete robotAccount")
	}
	return resp.StatusCode == http.StatusOK, nil
}

func (h *Client) CheckIfProjectExists(ctx context.Context, name string) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodHead,
		fmt.Sprintf("%s/projects", h.url.String()),
		nil,
	)
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

	return r2.StatusCode == http.StatusOK, nil
}

func (h *Client) CreateProject(ctx context.Context, name string) error {
	ok, err := h.CheckIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	body := map[string]any{
		"project_name": name,
		"public":       false,
	}
	bbody, err := json.Marshal(body)
	if err != nil {
		return errors.NewEf(err, "could not unmarshal req body")
	}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/%s", h.url.String(), "projects"),
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
	return errors.Newf("could not create Client project as received (statuscode=%d)", resp.StatusCode)
}

func (h *Client) DeleteProject(ctx context.Context, name string) error {
	_, err := h.CheckIfProjectExists(ctx, name)
	if err != nil {
		return err
	}

	u, err := h.url.Parse(name)
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
	return errors.Newf("could not delete Client project as received (statuscode=%d)", resp.StatusCode)
}

type Args struct {
	HarborAdminUsername string
	HarborAdminPassword string
	HarborRegistryHost  string
	HarborApiVersion    *string
}

func NewClient(args Args) (*Client, error) {
	if args.HarborApiVersion == nil {
		args.HarborApiVersion = fn.New("v2.0")
	}
	u, err := url.Parse("https://" + args.HarborRegistryHost + "/api/" + *args.HarborApiVersion)
	if err != nil {
		return nil, errors.NewEf(err, "registryUrl is not a valid url")
	}
	return &Client{
		args: &args,
		url:  u,
	}, nil
}
