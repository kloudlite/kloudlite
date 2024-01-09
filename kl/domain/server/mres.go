package server

import (
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Mres struct {
	DisplayName string   `json:"display_name"`
	Metadata    Metadata `json:"metadata"`
}

func ListMreses(options ...fn.Option) ([]Mres, error) {
	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	projectName, err := client.CurrentProjectName()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listMreses", map[string]any{
		"projectName": projectName,
		"envName":     env.Name,
		"pq":          PaginationDefault,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	fromResp, err := GetFromRespForEdge[Mres](respData)
	if err != nil {
		return nil, err
	}

	return fromResp, nil
}

func SelectMres(options ...fn.Option) (*Mres, error) {

	mresName := fn.GetOption(options, "mresName")

	m, err := ListMreses(options...)
	if err != nil {
		return nil, err
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
		return item.DisplayName
	}, fzf.WithPrompt("Select managed resource >"))

	return mres, err
}

func ListMresKeys(options ...fn.Option) ([]string, error) {
	mresName := fn.GetOption(options, "mresName")

	env, err := EnsureEnv(nil, options...)
	if err != nil {
		return nil, err
	}

	projectName, err := client.CurrentProjectName()
	if err != nil {
		return nil, err
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getMresKeys", map[string]any{
		"projectName": projectName,
		"envName":     env.Name,
		"name":        mresName,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	s, err := GetFromResp[[]string](respData)
	if err != nil {
		return nil, err
	}

	return *s, nil
}

func SelectMresKey(options ...fn.Option) (*string, error) {
	mresName := fn.GetOption(options, "mresName")

	keys, err := ListMresKeys(options...)
	if err != nil {
		return nil, err
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

	projectName := fn.GetOption(options, "projectName")
	projectName, err := EnsureProject(options...)

	if err != nil {
		return nil, err
	}

	env, err := EnsureEnv(&client.Env{
		Name: fn.GetOption(options, "envName"),
	}, options...)

	if err != nil {
		return nil, err
	}

	kt, err := client.GetKlFile(nil)
	if err != nil {
		return nil, err
	}

	if kt.Mres == nil {
		return nil, fmt.Errorf("no managed resource selected")
	}

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_getMresKeys", map[string]any{
		"projectName": projectName,
		"envName":     env.Name,
		"keyrefs": func() []map[string]string {
			var keyrefs []map[string]string
			for _, m := range kt.Mres {
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
		return nil, err
	}

	fromResp, err := GetFromResp[[]MresResp](respData)
	if err != nil {
		return nil, err
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

	for _, rt := range kt.Mres {
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
