package lib

import (
	"fmt"
	"io/ioutil"
	"kloudlite.io/cmd/internal/lib/server"
	"kloudlite.io/pkg/errors"
	"log"
	"os"
	"os/exec"
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

const loginUrl = "https://auth.local.kl.madhouselabs.io/cli-login"

func Login() {
	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		fmt.Println(err)
		return
	}
	command := exec.Command("open", fmt.Sprintf("%s/%s%s", loginUrl, "?loginId=", loginId))
	err = command.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = server.Login(loginId)
	if err != nil {
		fmt.Println(err)
		return
	}
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
