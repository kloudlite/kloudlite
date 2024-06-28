package apiclient

import (
	"fmt"

	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
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
// 		return nil, functions.NewE(err)
// 	}
//
// 	cookie, err := getCookie()
// 	if err != nil {
// 		return nil, functions.NewE(err)
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
// 		return nil, functions.NewE(err)
// 	}
//
// 	if fromResp, err := GetFromResp[Env](respData); err != nil {
// 		return nil, functions.NewE(err)
// 	} else {
// 		return fromResp, nil
// 	}
// }

func ListEnvs(options ...fn.Option) ([]Env, error) {
	var err error
	_, err = EnsureAccount(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
	}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromRespForEdge[Env](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

func GetEnvironment(accountName, envName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, err
	}
	respData, err := klFetch("cli_getEnvironment", map[string]any{
		"name": envName,
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

func SelectEnv(envName string, options ...fn.Option) (*Env, error) {

	persistSelectedEnv := func(env fileclient.Env) error {
		err := fileclient.SelectEnv(env)
		if err != nil {
			return functions.NewE(err)
		}
		return nil
	}

	envs, err := ListEnvs(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	oldEnv, _ := fileclient.CurrentEnv()

	if envName != "" {
		for _, a := range envs {
			if a.Metadata.Name == envName {
				port := 0
				if oldEnv != nil {
					port = oldEnv.SSHPort
				}

				if err := persistSelectedEnv(fileclient.Env{
					Name:        a.Metadata.Name,
					SSHPort:     port,
					TargetNs:    a.Spec.TargetNamespace,
					ClusterName: a.ClusterName,
				}); err != nil {
					return nil, functions.NewE(err)
				}
				return &a, nil
			}
		}
		return nil, functions.Error("you don't have access to this account")
	}

	env, err := fzf.FindOne(
		envs,
		func(env Env) string {
			return fmt.Sprintf("%s (%s)", env.DisplayName, env.Metadata.Name)
		},
		fzf.WithPrompt("Select Environment > "),
	)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if err := persistSelectedEnv(fileclient.Env{
		Name:     env.Metadata.Name,
		TargetNs: env.Spec.TargetNamespace,
		SSHPort: func() int {
			if oldEnv == nil {
				return 0
			}
			return oldEnv.SSHPort
		}(),
		ClusterName: env.ClusterName,
	}); err != nil {
		return nil, functions.NewE(err)
	}

	return env, nil
}

func EnsureEnv(env *fileclient.Env, options ...fn.Option) (*fileclient.Env, error) {
	fc, err := fileclient.New()
	if err != nil {
		return nil, functions.NewE(err)
	}

	accountName := fn.GetOption(options, "accountName")
	if _, err := EnsureAccount(
		fn.MakeOption("accountName", accountName),
	); err != nil {
		return nil, functions.NewE(err)
	}

	if env != nil && env.Name != "" {
		return env, nil
	}

	env, _ = fileclient.CurrentEnv()

	if env != nil {
		return env, nil
	}

	kl, err := fc.GetKlFile("")
	if err != nil {
		return nil, functions.NewE(err)
	}

	if kl.DefaultEnv == "" {
		return nil, functions.Error("please select an environment using 'kl use env'")
	}
	selectedEnv, err := SelectEnv(kl.DefaultEnv, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}
	return &fileclient.Env{
		Name:     selectedEnv.DisplayName,
		TargetNs: selectedEnv.Metadata.Namespace,
	}, nil
}
