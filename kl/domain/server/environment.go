package server

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Env struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

func ListEnvs(options ...fn.Option) ([]Env, error) {
	accountName := fn.GetOption(options, "accountName")
	clusterName := fn.GetOption(options, "clusterName")

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

	projectName, err := client.CurrentProjectName()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"project": map[string]any{
			"type":  "name",
			"value": projectName,
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Env](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func SelectEnv(envName string) (*Env, error) {

	persistSelectedEnv := func(envName string) error {
		err := client.SelectEnv(envName)
		if err != nil {
			return err
		}
		return nil
	}

	envs, err := ListEnvs()
	if err != nil {
		return nil, err
	}

	if envName != "" {
		for _, a := range envs {
			if a.Metadata.Name == envName {
				if err := persistSelectedEnv(a.Metadata.Name); err != nil {
					return nil, err
				}
				return &a, nil
			}
		}
		return nil, errors.New("you don't have access to this account")
	}

	env, err := fzf.FindOne(
		envs,
		func(env Env) string {
			return fmt.Sprintf("%s (%s)", env.DisplayName, env.Metadata.Name)
		},
		fzf.WithPrompt("Select Environment > "),
	)

	if err != nil {
		return nil, err
	}

	if err := persistSelectedEnv(env.Metadata.Name); err != nil {
		return nil, err
	}

	return env, nil
}
