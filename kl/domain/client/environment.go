package client

import (
	"errors"
	"os"
)

func SelectEnv(envName Env) error {

	k, err := GetContextFile()
	if err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if k.SelectedEnvs == nil {
		k.SelectedEnvs = map[string]*Env{}
	}

	k.SelectedEnvs[dir] = &envName
	if err := WriteContextFile(*k); err != nil {
		return err
	}

	return nil

}

func CurrentEnv() (*Env, error) {
	file, err := GetContextFile()
	if err != nil {
		return nil, err
	}

	if file.SelectedEnvs == nil {
		return nil, errors.New("No selected environment")
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if file.SelectedEnvs[dir] == nil {
		return nil, errors.New("No selected environment")
	}

	return file.SelectedEnvs[dir], nil
}
