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
	sessionFile, err := os.Stat(path.Join(configFolder, ConfigFileName))
	if err != nil && os.IsNotExist(err) {
		return errors.New("not logged in")
	}
	if err != nil {
		return err
	}

	infraContextFile, _ := os.Stat(path.Join(configFolder, InfraContextsFileName))
	if infraContextFile != nil {
		if err := os.Remove(path.Join(configFolder, infraContextFile.Name())); err != nil {
			return err
		}
	}

	accountFile, _ := os.Stat(path.Join(configFolder, AccountContextsFileName))
	if accountFile != nil {
		if err := os.Remove(path.Join(configFolder, accountFile.Name())); err != nil {
			return err
		}
	}

	extraDataFile, _ := os.Stat(path.Join(configFolder, ExtraDataFileName))
	if extraDataFile != nil {
		if err := os.Remove(path.Join(configFolder, extraDataFile.Name())); err != nil {
			return err
		}
	}

	wgpidFile, _ := os.Stat(path.Join(configFolder, "wgpid"))
	if wgpidFile != nil {
		if err := os.Remove(path.Join(configFolder, wgpidFile.Name())); err != nil {
			return err
		}
	}

	return os.Remove(path.Join(configFolder, sessionFile.Name()))
	//return os.RemoveAll(configFolder)
}
