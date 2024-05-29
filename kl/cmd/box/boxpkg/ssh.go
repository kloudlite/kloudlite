package boxpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/adrg/xdg"
	cl "github.com/kloudlite/kl/domain/client"
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func getDomainFromPath(pth string) string {

	s := strings.ToLower(pth)

	s = strings.ReplaceAll(s, xdg.Home, "")
	s = strings.ReplaceAll(s, ":\\", "/")
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", ".")
	s = strings.ReplaceAll(s, "\\", ".")
	s = strings.Trim(s, ".")
	s = fmt.Sprintf("%s.local.khost.dev", s)

	return s
}

func (c *client) Ssh() error {
	defer c.spinner.Start("preparing to ssh")()

	cname := fn.ParseStringFlag(c.cmd, "name")
	if cname != "" {
		return c.doSSHWithCname(cname)
	}

	_, err := cl.GetKlFile("")
	if err != nil {
		if err == confighandler.ErrKlFileNotExists {
			conts, err := c.listContainer(map[string]string{
				CONT_MARK_KEY: "true",
			})
			if err != nil {
				return err
			}

			cnt, err := fzf.FindOne(conts, func(item Cntr) string {
				return fmt.Sprintf("%s (%s)", item.Name, item.Labels[CONT_PATH_KEY])
			}, fzf.WithPrompt("Select Container > "))
			if err != nil {
				return err
			}

			return c.doSSHWithCname(cnt.Name)

		}

		return err
	}

	cr, err := c.getContainer(map[string]string{
		CONT_NAME_KEY: c.containerName,
		CONT_PATH_KEY: c.cwd,
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr || (err == nil && c.containerName != cr.Name) {
		err := c.Start()

		if err != nil && err != errContainerNotStarted {
			return err
		}
	}

	localEnv, err := cl.CurrentEnv()
	if err != nil {
		return err
	}

	if localEnv.SSHPort == 0 {
		return fmt.Errorf("container not running")
	}

	c.spinner.Stop()
	c.ensureVpnConnected()

	count := 0

	for {

		if !cl.CheckPortAvailable(localEnv.SSHPort) {
			break
		}

		count++
		if count == 10 {
			return fmt.Errorf("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start")
		}

		time.Sleep(1 * time.Second)
	}

	c.spinner.Stop()
	command := exec.Command("ssh", fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(localEnv.SSHPort), "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no")

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	// command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err.Error())
	}
	return nil
}

func (c *client) doSSHWithCname(name string) error {
	cr, err := c.getContainer(map[string]string{
		CONT_NAME_KEY: name,
	})

	if err == notFoundErr {
		return fmt.Errorf("no container running with name %s", name)
	}

	if err != nil {
		return err
	}

	pth := cr.Labels[CONT_PATH_KEY]
	localEnv, err := cl.EnvOfPath(pth)
	if err != nil {
		return err
	}

	if localEnv.SSHPort == 0 {
		return fmt.Errorf("container not running")
	}

	c.spinner.Stop()
	c.ensureVpnConnected()

	count := 0
	for {
		if err := exec.Command("ssh", fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no", "--", "exit 0").Run(); err == nil {
			break
		}

		count++
		if count == 10 {
			return fmt.Errorf("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s", err)
		}
		time.Sleep(1 * time.Second)
	}

	c.spinner.Stop()
	command := exec.Command("ssh", fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no")

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err.Error())
	}
	return nil
}
