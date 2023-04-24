package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kloudlite.io/pkg/errors"
	"net/http"
	"strconv"
	"strings"
)

type Robot struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	ReadOnly bool   `json:"read_only"`
	Secret   string `json:"secret"`
}

type Permission string

const (
	PermissionPush Permission = "push"
	PermissionPull Permission = "pull"
)

func (h *Client) GetRobot(ctx context.Context, accountName string, robotId int) (*Robot, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/robots/%d", robotId), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type RobotResult struct {
		Name        string `json:"name"`
		Secret      string `json:"secret"`
		Id          int    `json:"id"`
		Permissions []struct {
			Namespace string `json:"namespace"`
			Kind      string `json:"kind"`
			Access    []struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
				Effect   string `json:"effect"`
			}
		}
	}

	if resp.StatusCode == http.StatusOK {
		var robot RobotResult
		if err := json.Unmarshal(body, &robot); err != nil {
			return nil, err
		}
		readOnly := true
		for _, p := range robot.Permissions {
			if p.Kind == "project" {
				for _, a := range p.Access {
					if a.Action == "push" && a.Resource == "repository" {
						readOnly = false
					}
				}
			}
		}

		if !strings.HasPrefix(robot.Name, fmt.Sprintf("robot-%s", accountName)) {
			return nil, errors.New(http.StatusText(http.StatusNotFound))
		}

		return &Robot{
			Id:       robot.Id,
			Name:     robot.Name,
			ReadOnly: readOnly,
			Secret:   robot.Secret,
		}, nil
	}

	return nil, errors.New(http.StatusText(resp.StatusCode))
}

func (h *Client) CreateRobot(ctx context.Context, accountName string, name string, description *string, readOnly bool) (*Robot, error) {
	type AccessReq struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
		Effect   string `json:"effect"`
	}
	type PermissionsReq struct {
		Namespace string      `json:"namespace"`
		Kind      string      `json:"kind"`
		Access    []AccessReq `json:"access"`
	}
	type RobotReq struct {
		Name        string           `json:"name"`
		Level       string           `json:"level"`
		Duration    int              `json:"duration"`
		Permissions []PermissionsReq `json:"permissions"`
		Description string           `json:"description"`
	}

	req := RobotReq{
		Name:     name,
		Level:    "project",
		Duration: -1,
		Description: func() string {
			if description != nil {
				return *description
			}
			return ""
		}(),
		Permissions: []PermissionsReq{
			{
				Namespace: string(accountName),
				Kind:      "project",
				Access: []AccessReq{
					{
						Resource: "repository",
						Action:   string(PermissionPull),
						Effect:   "allow",
					},
				},
			},
		},
	}

	if !readOnly {
		req.Permissions[0].Access = append(req.Permissions[0].Access, AccessReq{
			Resource: "repository",
			Action:   string(PermissionPush),
			Effect:   "allow",
		})
	}

	reqMarshal, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	httpRequest, err := h.NewAuthzRequest(ctx, http.MethodPost, "/robots", strings.NewReader(string(reqMarshal)))
	if err != nil {
		return nil, err
	}
	type RobotResult struct {
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Secret      string `json:"secret"`
		Permissions []struct {
			Namespace string `json:"namespace"`
			Kind      string `json:"kind"`
			Access    []struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
				Effect   string `json:"effect"`
			}
		}
	}

	resp, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusCreated {
		var result RobotResult
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		var resp *Robot
		readOnly := true
		for _, p := range result.Permissions {
			if p.Kind == "project" {
				for _, a := range p.Access {
					if a.Action == "push" && a.Resource == "repository" {
						readOnly = false
					}
				}
			}
		}
		resp = &Robot{
			Id:       result.Id,
			Name:     result.Name,
			ReadOnly: readOnly,
			Secret:   result.Secret,
		}
		return resp, nil
	}
	return nil, errors.New(string(body))
}

func (h *Client) GetRobots(ctx context.Context, projectId int, searchOpts ListOptions) ([]Robot, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, "/robots", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("q", fmt.Sprintf("Level=project,ProjectID=%d", projectId))
	q.Add(
		"sort", func() string {
			if searchOpts.Sort == "" {
				return "-id"
			}
			return searchOpts.Sort
		}(),
	)
	q.Add("page", strconv.FormatInt(searchOpts.Page, 10))
	q.Add("page_size", strconv.FormatInt(searchOpts.PageSize, 10))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	type RobotResult struct {
		Name        string `json:"name"`
		Secret      string `json:"secret"`
		Id          int    `json:"id"`
		Permissions []struct {
			Namespace string `json:"namespace"`
			Kind      string `json:"kind"`
			Access    []struct {
				Resource string `json:"resource"`
				Action   string `json:"action"`
				Effect   string `json:"effect"`
			}
		}
	}

	if resp.StatusCode == http.StatusOK {
		var results []RobotResult
		if err := json.Unmarshal(body, &results); err != nil {
			return nil, err
		}
		var resp []Robot
		for _, r := range results {
			readOnly := true
			for _, p := range r.Permissions {
				if p.Kind == "project" {
					for _, a := range p.Access {
						if a.Action == "push" && a.Resource == "repository" {
							readOnly = false
						}
					}
				}
			}
			resp = append(resp, Robot{
				Id:       r.Id,
				Name:     r.Name,
				ReadOnly: readOnly,
				Secret:   r.Secret,
			})
		}
		return resp, nil
	}

	return nil, errors.Newf("bad status code (%d) received, with error message, %s", resp.StatusCode, body)
}

func (h *Client) DeleteRobot(ctx context.Context, accountName string, robotId int) error {
	_, err := h.GetRobot(ctx, accountName, robotId)
	if err != nil {
		return err
	}
	req, err := h.NewAuthzRequest(ctx, http.MethodDelete, fmt.Sprintf("/robots/%d", robotId), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	return errors.Newf("bad status code (%d) received, with error message, %s", resp.StatusCode, body)
}
