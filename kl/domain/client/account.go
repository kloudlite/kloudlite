package client

import "errors"

func SelectAccount(accountName string) error {
	file, err := GetActiveContext()
	if err != nil {
		return err
	}

	file.AccountName = accountName

	if file.Name == "" {
		return nil
	}

	err = WriteContextFile(*file)
	return err
}

func CurrentAccountName() (string, error) {
	file, err := GetActiveContext()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New("noSelectedContext")
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
