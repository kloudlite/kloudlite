package apiclient

import (
	"strings"

	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type ConfigORSecret struct {
	Entries map[string]string `json:"entries"`
	Name    string            `json:"name"`
}

type Config struct {
	DisplayName string            `yaml:"displayName"`
	Metadata    Metadata          `yaml:"metadata"`
	Status      Status            `yaml:"status"`
	Data        map[string]string `yaml:"data"`
}

func ListConfigs(options ...fn.Option) ([]Config, error) {

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listConfigs", map[string]any{
		"pq": map[string]any{
			"orderBy":       "name",
			"sortDirection": "ASC",
			"first":         99999999,
		},
		"envName": strings.TrimSpace(env.Name),
	}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromRespForEdge[Config](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}

func SelectConfig(options ...fn.Option) (*Config, error) {

	e, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	if e.Name == "" {
		return nil, functions.Error("no environment selected")
	}

	configs, err := ListConfigs(options...)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if len(configs) == 0 {
		return nil, functions.Error("no configs found")
	}

	config, err := fzf.FindOne(
		configs,
		func(config Config) string {
			return config.DisplayName
		},
	)

	if err != nil {
		return nil, functions.NewE(err)
	}

	return config, nil
}

func EnsureConfig(options ...fn.Option) (*Config, error) {
	configName := fn.GetOption(options, "configName")

	if configName != "" {
		return GetConfig(options...)
	}

	config, err := SelectConfig(options...)

	if err != nil {
		return nil, functions.NewE(err)
	}

	return config, nil
}

func GetConfig(options ...fn.Option) (*Config, error) {
	configName := fn.GetOption(options, "configName")

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}
	cookie, err := getCookie()
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_getConfig", map[string]any{
		"name":    configName,
		"envName": strings.TrimSpace(env.Name),
	}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}

	if fromResp, err := GetFromResp[Config](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return fromResp, nil
	}
}
