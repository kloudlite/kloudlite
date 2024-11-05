package apiclient

import (
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type Secret struct {
	DisplayName string            `yaml:"displayName" json:"displayName"`
	Metadata    Metadata          `yaml:"metadata" json:"metadata"`
	Status      Status            `yaml:"status" json:"status"`
	StringData  map[string]string `yaml:"stringData" json:"stringData"`
	IsReadyOnly bool              `yaml:"isReadyOnly" json:"isReadyOnly"`
}

func (apic *apiClient) ListSecrets(teamName string, envName string) ([]Secret, error) {

	// env, err := EnsureEnv(nil, options...)
	// if err != nil {
	// 	return nil, functions.NewE(err)
	// }

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_listSecrets", map[string]any{
		"pq": map[string]any{
			"orderBy":       "updateTime",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"envName": strings.TrimSpace(envName),
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}
	if fromResp, err := GetFromRespForEdge[Secret](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		var secrets []Secret
		for _, s := range fromResp {
			if !s.IsReadyOnly {
				secrets = append(secrets, s)
			}
		}
		return secrets, nil
	}
}

// func SelectSecret(options ...fn.Option) (*Secret, error) {
// 	e, err := EnsureEnv(nil, options...)
// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if e.Name == "" {
// 		return nil, functions.Error("no environment selected")
// 	}

// 	secrets, err := ListSecrets(options...)

// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	if len(secrets) == 0 {
// 		return nil, functions.Error("no secret found")
// 	}

// 	secret, err := fzf.FindOne(
// 		secrets,
// 		func(sec Secret) string {
// 			return sec.DisplayName
// 		},
// 	)

// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	return secret, nil
// }

// func EnsureSecret(options ...fn.Option) (*Secret, error) {
// 	secName := fn.GetOption(options, "secretName")

// 	if secName != "" {
// 		return GetSecret(options...)
// 	}

// 	secret, err := SelectSecret(options...)

// 	if err != nil {
// 		return nil, functions.NewE(err)
// 	}

// 	return secret, nil
// }

func (apic *apiClient) GetSecret(teamName string, secretName string) (*Secret, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getSecret", map[string]any{
		"name":    secretName,
		"envName": strings.TrimSpace(currentEnv.Name),
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := GetFromResp[Secret](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}
