package boxpkg

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func getDomainFromPath(pth string) string {

	s := strings.ReplaceAll(pth, xdg.Home, "")

	s = strings.ToLower(s)

	s = strings.ReplaceAll(s, ":\\", "/")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", ".")
	s = strings.ReplaceAll(s, "\\", ".")
	s = strings.Trim(s, ".")
	s = fmt.Sprintf("%s.local.khost.dev", s)

	return s
}

func (c *client) Ssh() error {
	defer spinner.Client.UpdateMessage("trying to ssh into the box")()

	dir, _ := os.Getwd()
	if envclient.InsideBox() {
		return fn.Error("you are already in a devbox")
	}

	err := c.Start()
	if err != nil && errors.Is(err, UserCanceled) {
		fn.Log("Operation was canceled by the user")
		return nil
	} else if err != nil {
		if err2 := c.Stop(); err2 != nil {
			return fn.NewE(err2)
		}
		return fn.NewE(err)
	}

	cont, err := c.containerAtPath(dir)
	if err != nil {
		return fn.NewE(err)
	}

	port, err := strconv.Atoi(cont.Labels[SSH_PORT_KEY])
	if err != nil {
		return fn.NewE(err)
	}

	if err := c.waitForSshReady(port, cont.ID); err != nil {
		return fn.NewE(err)
	}

	spinner.Client.Pause()
	if err := sshclient.DoSSH(sshclient.SSHConfig{
		User:    "kl",
		Host:    getDomainFromPath(cont.Labels[CONT_PATH_KEY]),
		SSHPort: port,
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
	}); err != nil {
		return fn.NewE(err)
	}

	spinner.Client.Resume()
	return nil
}

func sshConf(host string, port int) sshclient.SSHConfig {
	return sshclient.SSHConfig{
		User:    "kl",
		Host:    host,
		SSHPort: port,
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
	}
}
