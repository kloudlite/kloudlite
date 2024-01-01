package server

import (
	"encoding/json"
	"fmt"

	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/util"
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
	Edges    []Region `json:"edges"`
}

func GetRegions(options ...common_util.Option) ([]Region, error) {

	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	accountId := common_util.GetOption(options, "accountId")
	if accountId == "" {
		accountId, err = util.CurrentAccountName()

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
		for _, r := range p.Edges {
			r.Provider = p.Name
			r.Region = r.Id
			r.Name = fmt.Sprintf("(%s) %s, %s", p.Provider, p.Name, r.Name)
			regions = append(regions, r)
		}

	}

	return regions, nil
}
