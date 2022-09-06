package lib

import (
	"fmt"
	"io/ioutil"
	"kloudlite.io/cmd/internal/lib/common"
	"os"
	"os/exec"
	"runtime"

	"kloudlite.io/cmd/internal/lib/server"
	"kloudlite.io/pkg/errors"
)

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

func Login() error {
	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%s%s", loginUrl, "?loginId=", loginId)
	err = open(link)

	if err != nil {
		return err
	}
	err = server.Login(loginId)
	if err != nil {
		return err
	}
	return nil
}

func Logout() error {
	configFolder, err := common.GetConfigFolder()
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
	configFolder, err := common.GetConfigFolder()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFolder+"/account", []byte(accountId), 0644)
}
func SelectProject(projectId string) error {
	configFolder, err := common.GetConfigFolder()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFolder+"/project", []byte(projectId), 0644)
}
