package client

import "errors"

func CurrentClusterName() (string, error) {
	file, err := GetContextFile()
	if err != nil {
		return "", err
	}
	if file.ClusterName == "" {
		return "", errors.New("noSelectedCluster")
	}
	if file.ClusterName == "" {
		return "",
			errors.New("no clusters is selected yet. please select one using \"kl use cluster\"")
	}
	return file.ClusterName, nil
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
