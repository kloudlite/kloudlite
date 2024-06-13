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
	ClusterName string   `json:"clusterName"`
	Spec        struct {
		TargetNamespace string `json:"targetNamespace"`
	} `json:"spec"`
}

type EnvList struct {
	Edges Edges[Env] `json:"edges"`
}

// func GetEnvironment(envName string) (*Env, error) {
// 	var err error
// 	projectName, err := EnsureProject()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	respData, err := klFetch("cli_getEnvironment", map[string]any{
// 		"projectName": strings.TrimSpace(projectName),
// 		"pq": map[string]any{
// 			"orderBy":       "name",
// 			"sortDirection": "ASC",
// 			"first":         99999999,
// 		},
// 	}, &cookie)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if fromResp, err := GetFromResp[Env](respData); err != nil {
// 		return nil, err
// 	} else {
// 		return fromResp, nil
// 	}
// }

func ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error

	_, err = EnsureAccount(options...)
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

	oldEnv, _ := client.CurrentEnv()

	if envName != "" {
		for _, a := range envs {
			if a.Metadata.Name == envName {
				port := 0
				if oldEnv != nil {
					port = oldEnv.SSHPort
				}

				if err := persistSelectedEnv(client.Env{
					Name:        a.Metadata.Name,
					SSHPort:     port,
					TargetNs:    a.Spec.TargetNamespace,
					ClusterName: a.ClusterName,
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
		Name:        env.Metadata.Name,
		TargetNs:    env.Spec.TargetNamespace,
		SSHPort:     oldEnv.SSHPort,
		ClusterName: env.ClusterName,
	}); err != nil {
		return nil, err
	}

	return env, nil
}

func EnsureEnv(env *client.Env, options ...fn.Option) (*client.Env, error) {

	accountName := fn.GetOption(options, "accountName")
	if _, err := EnsureAccount(
		fn.MakeOption("accountName", accountName),
	); err != nil {
		return nil, err
	}

	if env != nil && env.Name != "" {
		return env, nil
	}

	env, _ = client.CurrentEnv()

	if env != nil {
		return env, nil
	}

	kl, err := client.GetKlFile("")
	if err != nil {
		return nil, err
	}
	if kl.DefaultEnv == "" {
		return nil, errors.New("please select an environment using 'kl use env'")
	}
	selectedEnv, err := SelectEnv(kl.DefaultEnv)
	if err != nil {
		return nil, err
	}
	return &client.Env{
		Name:     selectedEnv.DisplayName,
		TargetNs: selectedEnv.Metadata.Namespace,
	}, nil
}
