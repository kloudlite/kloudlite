package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"operators.kloudlite.io/lib/errors"
)

type Project struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type harborProject struct {
	UpdateTime         time.Time `json:"update_time"`
	OwnerName          string    `json:"owner_name"`
	Name               string    `json:"name"`
	Deleted            bool      `json:"deleted"`
	OwnerId            int       `json:"owner_id"`
	RepoCount          int       `json:"repo_count"`
	ChartCount         int       `json:"chart_count"`
	CreationTime       time.Time `json:"creation_time"`
	Togglable          bool      `json:"togglable"`
	CurrentUserRoleId  int       `json:"current_user_role_id"`
	CurrentUserRoleIds []int     `json:"current_user_role_ids"`
	CveAllowlist       struct {
		UpdateTime time.Time `json:"update_time"`
		Items      []struct {
			CveId string `json:"cve_id"`
		} `json:"items"`
		ProjectId    int       `json:"project_id"`
		CreationTime time.Time `json:"creation_time"`
		Id           int       `json:"id"`
		ExpiresAt    int       `json:"expires_at"`
	} `json:"cve_allowlist"`
	ProjectId  int `json:"project_id"`
	RegistryId int `json:"registry_id"`
	Metadata   struct {
		EnableContentTrust   string `json:"enable_content_trust"`
		Public               string `json:"public"`
		AutoScan             string `json:"auto_scan"`
		Severity             string `json:"severity"`
		RetentionId          string `json:"retention_id"`
		ReuseSysCveAllowlist string `json:"reuse_sys_cve_allowlist"`
		PreventVul           string `json:"prevent_vul"`
	} `json:"metadata"`
}

func (h *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	// ok, err := h.CheckIfProjectExists(ctx, name)
	// if err != nil {
	// 	return nil, err
	// }
	// if ok {
	// 	return h.GetProject(ctx, name)
	// }

	body := map[string]any{
		"project_name":  name,
		"public":        false,
		"storage_limit": 0, // unlimited space
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, errors.NewEf(err, "could not unmarshal req body")
	}

	req, err := h.NewAuthzRequest(ctx, http.MethodPost, "/projects", bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.NewEf(err, "could not build request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, errors.NewEf(err, "while making request")
	}
	if resp.StatusCode == http.StatusCreated {
		return &Project{
			Name:     name,
			Location: resp.Header.Get("Location"),
		}, nil
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, errors.Newf("could not create Client project as received (statuscode=%d), with error message, %s", resp.StatusCode, msg)
}

func (h *Client) GetProject(ctx context.Context, name string) (*Project, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/projects/%s", name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("bad status code (%d) received, with error msg, %s", resp.StatusCode, msg)
	}

	var p harborProject
	if err := json.Unmarshal(msg, &p); err != nil {
		return nil, err
	}

	return &Project{
		Name:     name,
		Location: fmt.Sprintf("/api/%s/projects/%d", *h.args.HarborApiVersion, p.ProjectId),
	}, nil
}

func (h *Client) CheckIfProjectExists(ctx context.Context, projectName string) (bool, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodHead, "/projects", nil)
	if err != nil {
		return false, err
	}
	q := req.URL.Query()
	q.Add("project_name", projectName)
	req.URL.RawQuery = q.Encode()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}

	return resp.StatusCode == http.StatusOK, nil
}

func (h *Client) SetProjectQuota(ctx context.Context, name string, storageSize int) error {
	body := map[string]any{
		"storage_limit": int64(storageSize),
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := h.NewAuthzRequest(ctx, http.MethodPut, fmt.Sprintf("/projects/%s", name), bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Newf("bad status-code=%d", resp.StatusCode)
	}
	return nil
}

func (h *Client) DeleteProject(ctx context.Context, name string) error {
	ok, err := h.CheckIfProjectExists(ctx, name)
	if err != nil {
		return err
	}
	if !ok {
		// ASSERt: project does not exist to be deleted
		return nil
	}

	req, err := h.NewAuthzRequest(ctx, http.MethodDelete, fmt.Sprintf("/projects/%s", name), nil)
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
	if resp.StatusCode == http.StatusNotFound {
		// ASSERt: silent exit, as harbor project already does not exist
		return nil
	}

	return errors.Newf("could not delete project=%s as received (statuscode=%d)", name, resp.StatusCode)
}
