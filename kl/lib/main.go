package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/common/ui/color"
	"github.com/kloudlite/kl/lib/server"
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

func Login() error {
	loginId, err := server.CreateRemoteLogin()
	if err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%s%s", constants.LoginUrl, "?loginId=", loginId)

	// count := 0

	// for {
	// 	fmt.Println(color.Text(fmt.Sprintf("Color %d", count), count))
	//
	// 	count++
	//
	// 	if count > 255 {
	// 		break
	// 	}
	// }

	fmt.Println(color.Text("Opening browser for login in the browser to authenticate your account\n", 2))
	fmt.Println(color.Text(link, 21))
	fmt.Println("")

	// if err = open(link); err != nil {
	// 	return err
	// }

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

func SelectAccount(accountId string) error {
	file, err := server.GetContextFile()
	if err != nil {
		return err
	}

	file.AccountId = accountId

	err = server.WriteContextFile(*file)
	return err
}

func SelectProject(projectId string) error {

	file, err := server.GetContextFile()
	if err != nil {
		return err
	}

	file.ProjectId = projectId

	err = server.WriteContextFile(*file)
	return err

}

func SelectDevice(deviceId string) error {

	file, err := server.GetContextFile()
	if err != nil {
		return err
	}

	file.DeviceId = deviceId

	err = server.WriteContextFile(*file)
	return err

}
