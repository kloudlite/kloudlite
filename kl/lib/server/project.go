package server

import (
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/util"
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
		accountName, err = util.CurrentAccountName()

		if err != nil {
			return nil, err
		}
	}

	if clusterName == "" {
		clusterName, err = util.CurrentClusterName()

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
