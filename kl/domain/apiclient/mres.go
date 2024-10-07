package apiclient

import (
	fn "github.com/kloudlite/kl/pkg/functions"
)

type Mres struct {
	DisplayName   string   `json:"displayName"`
	Name          string   `json:"name"`
	SecretRefName Metadata `json:"secretRef"`
}

func (apic *apiClient) ListMreses(teamName string, envName string) ([]Mres, error) {

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
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

func (apic *apiClient) ListMresKeys(teamName, envName, importedManagedResource string) ([]string, error) {
	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
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

func (apic *apiClient) GetMresConfigValues(teamName string) (map[string]string, error) {
	fc := apic.fc

	currentEnv, err := apic.EnsureEnv()
	if err != nil {
		return nil, fn.NewE(err)
	}

	kt, err := fc.GetKlFile("")
	if err != nil {
		return nil, fn.NewE(err)
	}

	if kt.EnvVars.GetMreses() == nil {
		return nil, fn.Errorf("no managed resource selected")
	}

	cookie, err := getCookie(fn.MakeOption("teamName", teamName))
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
