package client

import "errors"

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
		return "", errors.New("no account selected, please select one using \"kl use account\"")
	}
	if file.AccountName == "" {
		return "",
			errors.New("no context is selected yet. please select one using \"kl context switch\"")
	}
	return file.AccountName, nil
}
