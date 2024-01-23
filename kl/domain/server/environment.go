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
		TargetNamespace string `json:"targetNamespace"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

func GetEnvironment(envName string) (*Env, error) {
	var err error
	projectName, err := EnsureProject()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getEnvironment", map[string]any{
		"projectName": strings.TrimSpace(projectName),
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Env](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error
	projectName, err := EnsureProject(options...)
	if err != nil {
		return nil, err
	}
	if projectName == "" {
		return nil, fmt.Errorf("Please select a project using 'kl init'")
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"projectName": strings.TrimSpace(projectName),
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
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

func SelectEnv(envName string, options ...fn.Option) (*Env, error) {
	persistSelectedEnv := func(env client.Env) error {
		err := client.SelectEnv(env)
		if err != nil {
			return err
		}
		return nil
	}

	envs, err := ListEnvs(options...)
	if err != nil {
		return nil, err
	}

	if envName != "" {
		for _, a := range envs {
			if a.Metadata.Name == envName {
				if err := persistSelectedEnv(client.Env{
					Name:     a.Metadata.Name,
					TargetNs: a.Spec.TargetNamespace,
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
		Name:     env.Metadata.Name,
		TargetNs: env.Spec.TargetNamespace,
	}); err != nil {
		return nil, err
	}

	return env, nil
}

func EnsureEnv(env *client.Env, options ...fn.Option) (*client.Env, error) {
	if _, err := EnsureProject(options...); err != nil {
		return nil, err
	}

	if env != nil && env.Name != "" {
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
		Name:     mEnv.Metadata.Name,
		TargetNs: mEnv.Spec.TargetNamespace,
	}, nil
}
