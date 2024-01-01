package util

import "errors"

func CurrentProjectName() (string, error) {

	file, err := GetContextFile()

	if err != nil {
		return "", err
	}

	if file.ProjectId == "" {
		return "",
			errors.New("no project is selected yet. please select one using \"kl use project\"")
	}

	return file.ProjectId, nil
}
