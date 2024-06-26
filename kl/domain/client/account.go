package client

import fn "github.com/kloudlite/kl/pkg/functions"

func SelectAccount(accountName string) error {
	file, err := GetMainCtx()
	if err != nil {
		return err
	}

	file.AccountName = accountName

	if file.AccountName == "" {
		return nil
	}

	err = SetAccountToMainCtx(accountName)
	return err
}

func CurrentAccountName() (string, error) {

	kt, err := GetKlFile("")
	if err != nil {
		return "", err
	}

	if kt.AccountName == "" {
		return "", fn.Error("no account selected")
	}

	return kt.AccountName, nil
}
