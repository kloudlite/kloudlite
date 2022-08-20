package harbor

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

// Repository created by pasting json from harbor instance network tab
type Repository struct {
	ArtifactCount int       `json:"artifact_count"`
	CreationTime  time.Time `json:"creation_time"`
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	ProjectId     int       `json:"project_id"`
	PullCount     int       `json:"pull_count"`
	UpdateTime    time.Time `json:"update_time"`
}

func (h *Client) SearchRepositories(ctx context.Context, accountId repos.ID, searchQ string) ([]Repository, error) {
	req, err := h.NewAuthzRequest(ctx, http.MethodGet, fmt.Sprintf("/projects/%s/repositories", accountId), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("q", fmt.Sprintf("name=~%s", searchQ))
	q.Add("page", "1")
	q.Add("page_size", "100")
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var results []Repository
		if err := json.Unmarshal(body, &results); err != nil {
			return nil, err
		}
		return results, nil
	}

	return nil, errors.Newf("bad status code (%d) received, with error message, %s", resp.StatusCode, body)
}
