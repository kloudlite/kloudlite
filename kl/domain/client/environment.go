package client

import (
	"errors"
	"os"
)

func SelectEnv(envName Env) error {
	k, err := GetExtraData()
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

	return SaveExtraData(k)
}

func CurrentEnv() (*Env, error) {
	c, err := GetExtraData()
	if err != nil {
		return nil, err
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if c.SelectedEnvs == nil {
		return nil, errors.New("No selected environment")
	}

	if c.SelectedEnvs[dir] == nil {
		return nil, errors.New("No selected environment")
	}

	return c.SelectedEnvs[dir], nil
}
