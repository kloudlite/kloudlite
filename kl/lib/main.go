package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/constants"
	common "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/ui/text"
	"github.com/kloudlite/kl/lib/util"
)

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

	fmt.Println("opening browser for login")
	fmt.Println("if browser doesn't open automatically, please visit the following link")
	fmt.Println(url)

	return exec.Command(cmd, args...).Start()
}

func WhoAmI() error {
	if u, err := server.GetCurrentUser(); err != nil {
		return err
	} else {
		fmt.Println("You are logged in as " + text.Colored(u.Name, 4) + " (" + text.Colored(u.Email, 4) + ")")
		return nil
	}
}

func Login() error {
	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

	fmt.Println(text.Colored("Opening browser for login in the browser to authenticate your account\n", 2))
	fmt.Println(text.Colored(link, 21))
	fmt.Println("")

	if err = server.Login(loginId); err != nil {
		return err
	}
	return nil
}

func Logout() error {
	configFolder, err := common.GetConfigFolder()
	if err != nil {
		return err
	}
	_, err = os.Stat(configFolder + "/config")
	if err != nil && os.IsNotExist(err) {
		return errors.New("not logged in")
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(configFolder)
}

func SelectAccount(accountName string) error {
	file, err := util.GetContextFile()
	if err != nil {
		return err
	}

	file.AccountName = accountName

	err = util.WriteContextFile(*file)
	return err
}

func SelectCluster(clusterName string) error {
	file, err := util.GetContextFile()
	if err != nil {
		return err
	}

	file.ClusterName = clusterName

	err = util.WriteContextFile(*file)
	return err
}

func SelectProject(projectId string) error {

	file, err := util.GetContextFile()
	if err != nil {
		return err
	}

	file.ProjectId = projectId

	err = util.WriteContextFile(*file)
	return err

}

func SelectDevice(deviceId string) error {

	file, err := util.GetContextFile()
	if err != nil {
		return err
	}

	file.DeviceId = deviceId

	err = util.WriteContextFile(*file)
	return err

}
