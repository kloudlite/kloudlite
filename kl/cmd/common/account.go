package common_cmd

import (
	"errors"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/lib"
	"github.com/ktr0731/go-fuzzyfinder"
)

func SelectAccount(args []string) (*ResourceData, error) {
	persistSelectedAcc := func(accName string) error {
		err := lib.SelectAccount(accName)
		if err != nil {
			return err
		}
		return nil
	}
	accountId := ""
	if len(args) >= 1 {
		accountId = args[0]
	}
	accounts, err := server.ListAccounts()
	if err != nil {
		return nil, err
	}

	if accountId != "" {
		for _, a := range accounts {
			if a.Metadata.Name == accountId {
				if err := persistSelectedAcc(a.Metadata.Name); err != nil {
					return nil, err
				}
				return &ResourceData{
					Name:        a.Metadata.Name,
					DisplayName: a.DisplayName,
				}, nil
			}
		}
		return nil, errors.New("you don't have access to this account")
	}

	selectedIndex, err := fuzzyfinder.Find(
		accounts,
		func(i int) string {
			return accounts[i].DisplayName
		},
		fuzzyfinder.WithPromptString("Select Account > "),
	)

	if err != nil {
		return nil, err
	}
	if err := persistSelectedAcc(accounts[selectedIndex].Metadata.Name); err != nil {
		return nil, err
	}

	return &ResourceData{
		Name:        accounts[selectedIndex].Metadata.Name,
		DisplayName: accounts[selectedIndex].DisplayName,
	}, nil
}
