package apiclient

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
)

type Mres struct {
	DisplayName string   `json:"displayName"`
	Metadata    Metadata `json:"metadata"`
}

func (apic *apiClient) ListMreses(envName string, options ...fn.Option) ([]Mres, error) {

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_listImportedManagedResources", map[string]any{
		"envName": envName,
		"pq":      PaginationDefault,
	}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	fromResp, err := GetFromRespForEdge[Mres](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	return fromResp, nil
}

func (apic *apiClient) ListMresKeys(envName, importedManagedResource string, options ...fn.Option) ([]string, error) {
	cookie, err := getCookie(options...)
	if err != nil {
		return nil, fn.NewE(err)
	}
	respData, err := klFetch("cli_getSecret", map[string]any{
		"envName": envName,
		"name":    importedManagedResource,
	}, &cookie)
	if err != nil {
		return nil, fn.NewE(err)
	}

	s, err := GetFromResp[[]string](respData)
	if err != nil {
		return nil, fn.NewE(err)
	}

	return *s, nil
}

type MresResp struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	MresName string `json:"mresName"`
}

func (apic *apiClient) GetMresConfigValues(options ...fn.Option) (map[string]string, error) {
	fc := apic.fc

	// // env, err := EnsureEnv(&fileclient.Env{
	// // 	Name: fn.GetOption(options, "envName"),
	// // }, options...)

	// if err != nil {
	// 	return nil, fn.NewE(err)
	// }

	currentEnv, err := fc.CurrentEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}

	kt, err := fc.GetKlFile("")
	if err != nil {
		return nil, fn.NewE(err)
	}

	if kt.EnvVars.GetMreses() == nil {
		return nil, fmt.Errorf("no managed resource selected")
	}

	cookie, err := getCookie(options...)
	if err != nil {
		return nil, fn.NewE(err)
	}

	respData, err := klFetch("cli_getMresOutputKeyValues", map[string]any{
		"envName": currentEnv.Name,
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
		return nil, fn.NewE(err)
	}

	fromResp, err := GetFromResp[[]MresResp](respData)
	if err != nil {
		return nil, fn.NewE(err)
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
