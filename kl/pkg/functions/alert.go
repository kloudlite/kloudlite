package functions

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/martinlindhe/notify"
)

func Alert(name string, str ...interface{}) {
	if runtime.GOOS == "darwin" {
		notify.Alert("Kloudlite", name, fmt.Sprint(str...), "")
	}
	if runtime.GOOS == "linux" {
		notification(name, fmt.Sprint(str...), "")
		if err := exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/alarm-clock-elapsed.oga").Start(); err != nil {
			PrintError(NewE(err, "error playing alert sound"))
		}
	}
}

func Notify(name string, str ...interface{}) {
	if runtime.GOOS == "darwin" {
		notify.Notify("Kloudlite", name, fmt.Sprint(str...), "")
	}

	if runtime.GOOS == "linux" {
		notification(name, fmt.Sprint(str...), "")
	}
}

func notification(title string, txt string, iconPath string) {
	if euid := os.Geteuid(); euid == 0 {
		if usr, ok := os.LookupEnv("SUDO_USER"); ok {
			if euid, ok := os.LookupEnv("SUDO_UID"); ok {
				c := fmt.Sprintf("sudo -u %s DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/%s/bus notify-send -i %q %q %q", usr, euid, iconPath, title, txt)
				if err := ExecCmd(c, nil, false); err != nil {
					PrintError(err)
				}
			}
		}

		return
	}

	if err := ExecCmd(fmt.Sprintf("notify-send -i %q %q %q", iconPath, title, txt), nil, false); err != nil {
		PrintError(err)
	}
}
