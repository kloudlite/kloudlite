package apiclient

import (
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
)

type Account struct {
	Metadata    Metadata `json:"metadata"`
	DisplayName string   `json:"displayName"`
	Status      Status   `json:"status"`
}

func ListAccounts() ([]Account, error) {
	cookie, err := getCookie()
	if err != nil {
		return nil, functions.NewE(err)
	}

	respData, err := klFetch("cli_listAccounts", map[string]any{}, &cookie)
	if err != nil {
		return nil, functions.NewE(err)
	}

	type AccList []Account
	if fromResp, err := GetFromResp[AccList](respData); err != nil {
		return nil, functions.NewE(err)
	} else {
		return *fromResp, nil
	}
}

func SelectAccount(accountName string) (*Account, error) {

	accounts, err := ListAccounts()
	if err != nil {
		return nil, functions.NewE(err)
	}

	if accountName != "" {
		for _, a := range accounts {
			if a.Metadata.Name == accountName {
				return &a, nil
			}
		}
		return nil, functions.Error("you don't have access to this account")
	}

	account, err := fzf.FindOne(
		accounts,
		func(account Account) string {
			return account.DisplayName
		},
		fzf.WithPrompt("Select Account > "),
	)

	if err != nil {
		return nil, functions.NewE(err)
	}

	return account, nil
}

func EnsureAccount(options ...fn.Option) (string, error) {
	accountName := fn.GetOption(options, "accountName")

	if accountName != "" {
		return accountName, nil
	}

	s, _ := fileclient.CurrentAccountName()
	if s == "" {
		a, err := SelectAccount("")
		if err != nil {
			return "", functions.NewE(err)
		}

		return a.Metadata.Name, nil
	}

	return s, nil
}
