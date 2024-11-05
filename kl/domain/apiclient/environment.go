package apiclient

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
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

const (
	PublicEnvRoutingMode = "public"
	EnvironmentType      = "environment"
)

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

func (apic *apiClient) ListEnvs(teamName string) ([]Env, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listEnvironments", map[string]any{
		"pq": map[string]any{
			"orderBy":       "updateTime",
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

func (apic *apiClient) GetEnvironment(teamName, envName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
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

func (apic *apiClient) EnsureEnv() (*fileclient.Env, error) {
	CurrentEnv, err := apic.fc.CurrentEnv()
	if err != nil && err.Error() != fileclient.NoEnvSelected.Error() {
		return nil, functions.NewE(err)
	} else if err == nil {
		return CurrentEnv, nil
	}
	kt, err := apic.fc.GetKlFile("")
	if err != nil {
		return nil, functions.NewE(err)
	}
	if kt.DefaultEnv == "" {
		return nil, functions.Error("please initialize kl.yml by running `kl init` in current workspace")
	}
	e, err := apic.GetEnvironment(kt.TeamName, kt.DefaultEnv)
	if err != nil {
		return nil, functions.NewE(err)
	}
	return &fileclient.Env{
		Name:    e.DisplayName,
		SSHPort: 0,
	}, nil
}

// func _EnsureEnv(env *fileclient.Env, options ...fn.Option) (*fileclient.Env, error) {
// 	fc, err := fileclient.New()
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	teamName := fn.GetOption(options, "teamName")
// 	if _, err := EnsureTeam(
// 		fn.MakeOption("teamName", teamName),
// 	); err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if env != nil && env.Name != "" {
// 		return env, nil
// 	}

// 	env, _ = fc.CurrentEnv()

// 	if env != nil {
// 		return env, nil
// 	}

// 	kl, err := fc.GetKlFile("")
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if kl.DefaultEnv == "" {
// 		return nil, functions.Error("please select an environment using 'kl use env'")
// 	}
// 	selectedEnv, err := SelectEnv(kl.DefaultEnv, options...)
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}
// 	return &fileclient.Env{
// 		Name:     selectedEnv.DisplayName,
// 		TargetNs: selectedEnv.Metadata.Namespace,
// 	}, nil
// }

func (apic *apiClient) CloneEnv(teamName, envName, newEnvName, clusterName string) (*Env, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, functions.NewE(err)
	}
	respData, err := klFetch("cli_cloneEnvironment", map[string]any{
		"clusterName":            clusterName,
		"sourceEnvName":          envName,
		"destinationEnvName":     newEnvName,
		"displayName":            newEnvName,
		"environmentRoutingMode": PublicEnvRoutingMode,
	}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromResp[Env](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, functions.NewE(err)
	}
}

func (apic *apiClient) CheckEnvName(teamName, envName string) (bool, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return false, functions.NewE(err)
	}
	respData, err := klFetch("cli_coreCheckNameAvailability", map[string]any{
		"resType": EnvironmentType,
		"name":    envName,
	}, &cookie)
	if err != nil {
		return false, functions.NewE(err)
	}

	if fromResp, err := GetFromResp[CheckName](respData); err != nil {
		return false, functions.NewE(err)
	} else {
		return fromResp.Result, nil
	}
}
