package client

import (
	"errors"
	"os"
)

func SelectEnv(envName string) error {

	k, err := GetContextFile()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	k.SelectedEnvs[dir] = envName
	if err := WriteContextFile(*k); err != nil {
		return err
	}

	return nil

}

func CurrentEnvName() (string, error) {
	file, err := GetContextFile()
	if err != nil {
		return "", err
	}

	if file.SelectedEnvs == nil {
		return "", errors.New("noSelectedEnv")
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if file.SelectedEnvs[dir] == "" {
		return "", errors.New("noSelectedEnv")
	}

	return file.SelectedEnvs[dir], nil
}
