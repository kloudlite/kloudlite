package client

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func SelectAccount(accountName string) error {
	file, err := GetMainCtx()
	if err != nil {
		return functions.NewE(err)
	}

	file.AccountName = accountName

	if file.AccountName == "" {
		return nil
	}

	err = SetAccountToMainCtx(accountName)
	return functions.NewE(err)
}

func CurrentAccountName() (string, error) {

	kt, err := GetKlFile("")
	if err != nil {
		return "", functions.NewE(err)
	}

	if kt.AccountName == "" {
		return "", fn.Error("no account selected")
	}

	return kt.AccountName, nil
}
