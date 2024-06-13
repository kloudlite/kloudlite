package server

import (
	"errors"
	"fmt"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
)

type Account struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
}

func ListAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, err
	}

	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, err
	}

	type AccList []Account
	if fromResp, err := GetFromResp[AccList](respData); err != nil {
		return nil, err
	} else {
		return *fromResp, nil
	}
}

func SelectAccount(accountName string) (*Account, error) {
	persistSelectedAcc := func(accName string) error {
		err := client.SelectAccount(accName)
		if err != nil {
			return err
		}
		return nil
	}

	accounts, err := ListAccounts()
	if err != nil {
		return nil, err
	}

	if accountName != "" {
		for _, a := range accounts {
			if a.Metadata.Name == accountName {
				if err := persistSelectedAcc(a.Metadata.Name); err != nil {
					return nil, err
				}
				return &a, nil
			}
		}
		return nil, errors.New("you don't have access to this account")
	}

	account, err := fzf.FindOne(
		accounts,
		func(account Account) string {
			return account.DisplayName
		},
		fzf.WithPrompt("Select Account > "),
	)

	if err != nil {
		return nil, err
	}
	if err := persistSelectedAcc(account.Metadata.Name); err != nil {
		return nil, err
	}

	return account, nil
}

func EnsureAccount(options ...fn.Option) (string, error) {
	accountName := fn.GetOption(options, "accountName")

	returnAcc := func(an string) (string, error) {
		kt, err := client.GetKlFile("")
		if err != nil {
			return an, nil
		}

		if kt.AccountName != "" && kt.AccountName != an {
			return "", fmt.Errorf("selected account(%s) is not same as current workspace account (%s), please select account using '%s'", text.Yellow(an), text.Yellow(kt.AccountName), text.Bold("kl use account"))
		}

		if kt.AccountName == "" {
			kt.AccountName = an
			if err := client.WriteKLFile(*kt); err != nil {
				return "", err
			}
		}

		return an, nil
	}

	if accountName != "" {
		return returnAcc(accountName)
	}

	s, err := client.CurrentAccountName()
	if err != nil {
		return "", err
	}
	if s == "" {
		return "", errors.New(fmt.Sprintf("no account selected, please select an account using '%s use account'", flags.CliName))
	}

	return returnAcc(s)
}
