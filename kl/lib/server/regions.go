package server

import (
	"encoding/json"

	"github.com/kloudlite/kl/lib/common"
)

type Region struct {
	Id       string `json:"Id"`
	Name     string `json:"Name"`
	Region   string `json:"region"`
	Provider string
}

type Provider struct {
	Name     string   `json:"Name"`
	Provider string   `json:"provider"`
	Regions  []Region `json:"regions"`
}

func GetRegions(options ...common.Option) ([]Region, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	accountId := common.GetOption(options, "accountId")
	if accountId == "" {
		accountId, err = CurrentAccountId()

		if err != nil {
			return nil, err
		}
	}

	respData, err := klFetch("cli_ProviderWithRegions", map[string]any{
		"accountId": accountId,
	}, &cookie)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Providers []Provider `json:"data"`
	}

	var resp Response
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	regions := make([]Region, 0)
	for _, p := range resp.Providers {
		for _, r := range p.Regions {
			r.Provider = p.Name
			regions = append(regions, r)
		}

	}

	return regions, nil
}
