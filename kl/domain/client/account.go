package client

import "errors"

func SelectAccount(accountName string) error {
	file, err := GetContextFile()
	if err != nil {
		return err
	}

	file.AccountName = accountName

	err = WriteContextFile(*file)
	return err
}

func CurrentAccountName() (string, error) {
	file, err := GetContextFile()
	if err != nil {
		return "", err
	}
	if file.AccountName == "" {
		return "", errors.New("noSelectedCluster")
	}
	if file.AccountName == "" {
		return "",
			errors.New("no accounts is selected yet. please select one using \"kl use account\"")
	}
	return file.AccountName, nil
}
