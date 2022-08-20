package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"operators.kloudlite.io/lib/errors"
)

type Project struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

func (h *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	// ok, err := h.CheckIfProjectExists(ctx, name)
	// if err != nil {
	// 	return nil, err
	// }
	// if ok {
	// 	return h.getProject(ctx, name)
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

func (h *Client) getProject(ctx context.Context, name string) (*Project, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/projects/%s", name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return &Project{
			Name: name,
		}, nil
	}
	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return nil, errors.Newf("bad status code (%d) received, with error msg, %s", resp.StatusCode, msg)
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
