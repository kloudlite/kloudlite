package apiclient

import (
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
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

func (apic *apiClient) ListConfigs(accountName string, envName string) ([]Config, error) {

	// env, err := EnsureEnv(nil, options...)
	// if err != nil {
	// 	return nil, fn.NewE(err)
	// }

	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, fn.NewE(err)
	}
	//currentEnv, err := apic.fc.CurrentEnv()
	//if err != nil {
	//	return nil, fn.NewE(err)
	//}

	respData, err := klFetch("cli_listConfigs", map[string]any{
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

	if fromResp, err := GetFromRespForEdge[Config](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}

// func SelectConfig(options ...fn.Option) (*Config, error) {

// 	e, err := EnsureEnv(nil, options...)
// 	if err != nil {
// 		return nil, fn.NewE(err)
// 	}

// 	if e.Name == "" {
// 		return nil, fn.Error("no environment selected")
// 	}

// 	configs, err := ListConfigs(options...)

// 	if err != nil {
// 		return nil, fn.NewE(err)
// 	}

// 	if len(configs) == 0 {
// 		return nil, fn.Error("no configs found")
// 	}

// 	config, err := fzf.FindOne(
// 		configs,
// 		func(config Config) string {
// 			return config.DisplayName
// 		},
// 	)

// 	if err != nil {
// 		return nil, fn.NewE(err)
// 	}

// 	return config, nil
// }

// func EnsureConfig(options ...fn.Option) (*Config, error) {
// 	configName := fn.GetOption(options, "configName")

// 	if configName != "" {
// 		return GetConfig(options...)
// 	}

// 	config, err := SelectConfig(options...)

// 	if err != nil {
// 		return nil, fn.NewE(err)
// 	}

// 	return config, nil
// }

func (apic *apiClient) GetConfig(accountName string, envName string, configName string) (*Config, error) {

	cookie, err := getCookie(fn.MakeOption("accountName", accountName))
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getConfig", map[string]any{
		"name":    configName,
		"envName": envName,
	}, &cookie)

	if err != nil {
		return nil, fn.NewE(err)
	}

	if fromResp, err := GetFromResp[Config](respData); err != nil {
		return nil, fn.NewE(err)
	} else {
		return fromResp, nil
	}
}
