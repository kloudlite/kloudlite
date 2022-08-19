package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"kloudlite.io/cmd/internal/lib/server"
	"kloudlite.io/pkg/errors"
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

const loginUrl = "https://auth.kloudlite.io/cli-login"

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func Login() {
	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		fmt.Println(err)
		return
	}

	link := fmt.Sprintf("%s/%s%s", loginUrl, "?loginId=", loginId)
	err = open(link)

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

func Logout() error {
	configFolder, err := getConfigFolder()
	if err != nil {
		return err
	}
	_, err = os.Stat(configFolder + "/session")
	if err != nil && os.IsNotExist(err) {
		return errors.New("not logged in")
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(configFolder)
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
