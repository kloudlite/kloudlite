package boxpkg

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	cl "github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
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
	if klFile, err := cl.GetKlFile(""); err != nil {
		return fn.NewE(err)
	} else {
		dir, _ := os.Getwd()
		if os.Getenv("IN_DEV_BOX") == "true" {
			return fn.Error("you are already in a devbox")
		}

		c.SetCwd(dir)
		err = c.Start(klFile)
		if err != nil {
			if err2 := c.Stop(); err != nil {
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

		if err := c.waithForSshReady(port, cont.ID); err != nil {
			return fn.NewE(err)
		}

		fmt.Println("sshing into", getDomainFromPath(cont.Labels[CONT_PATH_KEY]))

		if err := sshclient.DoSSH(sshclient.SSHConfig{
			User:    "kl",
			Host:    getDomainFromPath(cont.Labels[CONT_PATH_KEY]),
			SSHPort: port,
			KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
		}); err != nil {
			return fn.NewE(err)
		}
	}

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
