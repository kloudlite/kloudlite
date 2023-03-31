package harbor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Project struct {
	ProjectId int `json:"project_id"`
}

func (h *Client) CreateProject(ctx context.Context, accountName string) (*Project, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodPost, "/projects", strings.NewReader(fmt.Sprintf(`
		{
			"project_name":"%s",
			"storage_limit":-1,
			"metadata": {
				"public": "false"
			}
		}
`, accountName)))
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("bad status code (%d)", resp.StatusCode)
	}
	projectLocation := resp.Header.Get("Location")
	if projectLocation == "" {
		return nil, fmt.Errorf("no project location")
	}
	// get project id from location
	parts := strings.Split(projectLocation, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid project location")
	}
	projectId := parts[len(parts)-1]
	var project Project
	// convert projectId to int
	if _, err := fmt.Sscanf(projectId, "%d", &project.ProjectId); err != nil {
		return nil, err
	}
	return &project, nil
}
