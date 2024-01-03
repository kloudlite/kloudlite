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

func SelectCluster(clusterName string) error {
	file, err := GetContextFile()
	if err != nil {
		return err
	}

	file.ClusterName = clusterName

	err = WriteContextFile(*file)
	return err
}
