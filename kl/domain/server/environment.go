package server

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Env struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
	Status      Status   `json:"status"`
	Spec        struct {
		IsEnvironment bool `json:"isEnvironment"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

func ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error
	projectName, err := EnsureProject(options...)
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
			"value": strings.TrimSpace(projectName),
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

	persistSelectedEnv := func(env client.Env) error {
		err := client.SelectEnv(env)
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
				if err := persistSelectedEnv(client.Env{
					Name:          a.Metadata.Name,
					IsEnvironment: a.Spec.IsEnvironment,
				}); err != nil {
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

	if err := persistSelectedEnv(client.Env{
		Name:          env.Metadata.Name,
		IsEnvironment: env.Spec.IsEnvironment,
	}); err != nil {
		return nil, err
	}

	return env, nil
}

func EnsureEnv(env *client.Env, options ...fn.Option) (*client.Env, error) {
	if _, err := EnsureProject(options...); err != nil {
		return nil, err
	}

	if env != nil {
		return env, nil
	}

	env, _ = client.CurrentEnv()

	if env != nil {
		return env, nil
	}

	mEnv, err := SelectEnv(func() string {
		if env != nil {
			return env.Name
		}
		return ""
	}())
	if err != nil {
		return nil, err
	}

	return &client.Env{
		Name:          mEnv.Metadata.Name,
		IsEnvironment: mEnv.Spec.IsEnvironment,
	}, nil
}
