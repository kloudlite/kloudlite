package server

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Mres struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
}

func ListMreses(options ...fn.Option) ([]Mres, error) {
	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listMreses", map[string]any{
		"envName": env.Name,
		"search": map[string]any{
			"envName": map[string]any{
				"matchType": "exact",
				"exact":     env.Name,
			},
		},
		"pq": PaginationDefault,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	fromResp, err := GetFromRespForEdge[Mres](respData)
	if err != nil {
		return nil, functions.NewE(err)
	}

	return fromResp, nil
}

func SelectMres(options ...fn.Option) (*Mres, error) {

	mresName := fn.GetOption(options, "mresName")

	m, err := ListMreses(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("no managed resources created yet on server")
	}

	if mresName != "" {
		for _, a := range m {
			if a.Metadata.Name == mresName {
				return &a, nil
			}
		}
		return nil, fmt.Errorf("you don't have access to this managed resource")
	}

	mres, err := fzf.FindOne(m, func(item Mres) string {
		return fmt.Sprintf("%s (%s)", item.DisplayName, item.Metadata.Name)
	}, fzf.WithPrompt("Select managed resource >"))

	return mres, err
}

func ListMresKeys(options ...fn.Option) ([]string, error) {
	mresName := fn.GetOption(options, "mresName")

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_getMresKeys", map[string]any{
		"envName": env.Name,
		"name":    mresName,
	}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	s, err := GetFromResp[[]string](respData)
	if err != nil {
		return nil, functions.NewE(err)
	}

	return *s, nil
}

func SelectMresKey(options ...fn.Option) (*string, error) {
	mresName := fn.GetOption(options, "mresName")

	keys, err := ListMresKeys(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys found in %s managed resource", mresName)
	}

	key, err := fzf.FindOne(keys, func(item string) string {
		return item
	}, fzf.WithPrompt("Select key >"))

	return key, err
}

type MresResp struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	MresName string `json:"mresName"`
}

func GetMresConfigValues(options ...fn.Option) (map[string]string, error) {
	env, err := EnsureEnv(&client.Env{
		Name: fn.GetOption(options, "envName"),
	}, options...)

	if err != nil {
		return nil, functions.NewE(err)
	}

	kt, err := client.GetKlFile("")
	if err != nil {
		return nil, functions.NewE(err)
	}

	if kt.EnvVars.GetMreses() == nil {
		return nil, fmt.Errorf("no managed resource selected")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_getMresOutputKeyValues", map[string]any{
		"envName": env.Name,
		"keyrefs": func() []map[string]string {
			var keyrefs []map[string]string
			for _, m := range kt.EnvVars.GetMreses() {
				for _, e := range m.Env {
					keyrefs = append(keyrefs, map[string]string{
						"mresName": m.Name,
						"key":      e.RefKey,
					})
				}
			}
			return keyrefs
		}(),
	}, &cookie)

	if err != nil {
		return nil, functions.NewE(err)
	}

	fromResp, err := GetFromResp[[]MresResp](respData)
	if err != nil {
		return nil, functions.NewE(err)
	}

	mresMap := map[string]map[string]*Kv{}

	for _, m := range *fromResp {
		if mresMap[m.MresName] == nil {
			mresMap[m.MresName] = map[string]*Kv{}
		}

		mresMap[m.MresName][m.Key] = &Kv{
			Key:   m.Key,
			Value: m.Value,
		}
	}

	result := map[string]string{}

	for _, rt := range kt.EnvVars.GetMreses() {
		for _, e := range rt.Env {

			if mresMap[rt.Name] == nil {
				continue
			}

			if mresMap[rt.Name][e.RefKey] == nil {
				continue
			}

			result[e.Key] = mresMap[rt.Name][e.RefKey].Value
		}
	}

	return result, nil
}
