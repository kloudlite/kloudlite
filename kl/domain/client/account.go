package client

import "errors"

func SelectAccount(accountName string) error {
	file, err := GetAccountContext()
	if err != nil {
		return err
	}

	file.AccountName = accountName

	if file.AccountName == "" {
		return nil
	}

	err = WriteAccountContext(accountName)
	return err
}

func CurrentAccountName() (string, error) {
	file, err := GetAccountContext()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New("no Selected account")
	}
	if file.AccountName == "" {
		return "",
			errors.New("no context is selected yet. please select one using \"kl context switch\"")
	}
	return file.AccountName, nil
}

func CurrentInfraAccountName() (string, error) {
	file, err := GetActiveInfraContext()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New("noSelectedInfraContext")
	}
	if file.AccountName == "" {
		return "",
			errors.New("no infra context is selected yet. please select one using \"kl infra context switch\"")
	}
	return file.AccountName, nil
}
