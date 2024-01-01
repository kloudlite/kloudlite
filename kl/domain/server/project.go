package server

import (
	"github.com/kloudlite/kl/domain/client"
	common_util "github.com/kloudlite/kl/pkg/functions"
)

type Project struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
}

type ProjectList struct {
	Edges Edges[Project] `json:"edges"`
}

func ListProjects(options ...common_util.Option) ([]Project, error) {
	accountName := common_util.GetOption(options, "accountName")
	clusterName := common_util.GetOption(options, "clusterName")

	var err error

	if accountName == "" {
		accountName, err = client.CurrentAccountName()

		if err != nil {
			return nil, err
		}
	}

	if clusterName == "" {
		clusterName, err = client.CurrentClusterName()

		if err != nil {
			return nil, err
		}
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listProjects", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Project](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}
