package server

import (
	"errors"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Secret struct {
	DisplayName string            `yaml:"displayName"`
	Metadata    Metadata          `yaml:"metadata"`
	Status      Status            `yaml:"status"`
	StringData  map[string]string `yaml:"stringData"`
}

func ListSecrets(options ...fn.Option) ([]Secret, error) {

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	projectName, err := client.CurrentProjectName()

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listSecrets", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"projectName": strings.TrimSpace(projectName),
		"envName":     strings.TrimSpace(env.Name),
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromRespForEdge[Secret](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}

func SelectSecret(options ...fn.Option) (*Secret, error) {

	e, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	if e.Name == "" {
		return nil, errors.New("no environment selected")
	}

	secrets, err := ListSecrets(options...)

	if err != nil {
		return nil, err
	}

	if len(secrets) == 0 {
		return nil, errors.New("no configs found")
	}

	secret, err := fzf.FindOne(
		secrets,
		func(sec Secret) string {
			return sec.DisplayName
		},
	)

	if err != nil {
		return nil, err
	}

	return secret, nil
}

func EnsureSecret(options ...fn.Option) (*Secret, error) {
	secName := fn.GetOption(options, "secretName")

	if secName != "" {
		return GetSecret(options...)
	}

	secret, err := SelectSecret(options...)

	if err != nil {
		return nil, err
	}

	return secret, nil
}

func GetSecret(options ...fn.Option) (*Secret, error) {
	secName := fn.GetOption(options, "secretName")

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	projectName, err := client.CurrentProjectName()

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getSecret", map[string]any{
		"name":        secName,
		"projectName": strings.TrimSpace(projectName),
		"envName":     strings.TrimSpace(env.Name),
	}, &cookie)

	if err != nil {
		return nil, err
	}

	if fromResp, err := GetFromResp[Secret](respData); err != nil {
		return nil, err
	} else {
		return fromResp, nil
	}
}
