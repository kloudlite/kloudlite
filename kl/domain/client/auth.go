package client

import (
	"errors"
	"os"
	"path"
)

func Logout() error {
	configFolder, err := GetConfigFolder()
	if err != nil {
		return err
	}
	_, err = os.Stat(path.Join(configFolder, ConfigFileName))
	if err != nil && os.IsNotExist(err) {
		return errors.New("not logged in")
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(configFolder)
}
