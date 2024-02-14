package client

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/flags"
)

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
	file, err := GetMainCtx()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New(fmt.Sprintf("no account selected, please select an account using '%s use account'", flags.CliName))
	}
	return file.AccountName, nil
}
