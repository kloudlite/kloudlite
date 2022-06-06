package lib

import (
	"fmt"
	"io/ioutil"
	"kloudlite.io/cmd/internal/lib/server"
	"kloudlite.io/pkg/errors"
	"log"
	"os"
)

func getConfigFolder() (configFolder string, err error) {
	var dirName string
	dirName, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if !ok {
		dirName, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}
	configFolder = fmt.Sprintf("%s/.kl", dirName)
	if _, err := os.Stat(configFolder); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(configFolder, os.ModePerm)
		if err != nil {
			log.Println(err)
		}
	}
	return configFolder, nil
}

func CheckLogin() error {
	_, err := server.Me()
	if err != nil {
		return err
	}
	return nil
}

func Login(authToken string) error {
	return ioutil.WriteFile("/tmp/auth.token", []byte(authToken), 0644)
}

func SelectAccount(accountId string) error {
	configFolder, err := getConfigFolder()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFolder+"/account", []byte(accountId), 0644)
}

func SelectProject(projectId string) error {
	configFolder, err := getConfigFolder()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFolder+"/project", []byte(projectId), 0644)
}
