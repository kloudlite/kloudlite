package server

import (
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

type App struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
}

func ListApps(options ...fn.Option) ([]App, error) {

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	projectName, err := client.CurrentProjectName()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listApps", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"project": map[string]any{
			"type":  "name",
			"value": strings.TrimSpace(projectName),
		},
		"scope": map[string]any{
			"type": func() string {
				if env.IsEnvironment {
					return "environmentName"
				}

				return "workspaceName"
			}(),
			"value": strings.TrimSpace(env.Name),
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[App](respData); err != nil {
		return nil, err
	} else {
		fmt.Println(fromResp)
		return fromResp, nil
	}
}
