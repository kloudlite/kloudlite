package functions

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenUrl(url string) error {
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

	Log("opening browser for login")
	Log("if browser doesn't open automatically, please visit the following link")
	fmt.Println(url)

	return exec.Command(cmd, args...).Start()
}
