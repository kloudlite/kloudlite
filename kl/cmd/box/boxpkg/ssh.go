package boxpkg

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/adrg/xdg"
	cl "github.com/kloudlite/kl/domain/client"
	confighandler "github.com/kloudlite/kl/pkg/config-handler"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/fzf"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
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
	defer spinner.Client.Start("preparing to ssh")()

	cname := fn.ParseStringFlag(c.cmd, "name")
	if cname != "" {
		return c.doSSHWithCname(cname)
	}

	_, err2 := cl.GetKlFile("")
	if err2 != nil {
		if err2 == confighandler.ErrKlFileNotExists {
			conts, err := c.listContainer(map[string]string{
				CONT_MARK_KEY: "true",
			})

			if err != nil && err == notFoundErr {
				return fmt.Errorf("kl.yml in current dir not found, and also no any running container found")
			}

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

		return err2
	}

	cr, err := c.getContainer(map[string]string{
		CONT_NAME_KEY: c.containerName,
		CONT_PATH_KEY: c.cwd,
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr || (err == nil && c.containerName != cr.Name) || (cr.State == ContStateExited || cr.State == ContStateCreated) {
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

	spinner.Client.Stop()
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

	spinner.Client.Stop()
	// command := exec.Command("ssh", fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(localEnv.SSHPort), "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no")

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	if err := sshclient.DoSSH(sshclient.SSHConfig{
		Host:    getDomainFromPath(c.cwd),
		User:    "kl",
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
		SSHPort: localEnv.SSHPort,
	}); err != nil {
		return err
	}

	// // command.Stderr = os.Stderr
	// command.Stdin = os.Stdin
	// command.Stdout = os.Stdout
	// if err := command.Run(); err != nil {
	// 	return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err.Error())
	// }
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

	if cr.State == ContStateExited || cr.State == ContStateCreated {
		if err := c.StopCont(cr); err != nil {
			return err
		}

		fn.PrintError(fmt.Errorf("container was not in running, stopped it, please try again"))
		return nil
	}

	pth := cr.Labels[CONT_PATH_KEY]
	localEnv, err := cl.EnvOfPath(pth)
	if err != nil {
		return err
	}

	if localEnv.SSHPort == 0 {
		return fmt.Errorf("container not running")
	}

	spinner.Client.Stop()
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

	spinner.Client.Stop()
	// command := exec.Command("ssh", fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no")

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	if err := sshclient.DoSSH(sshclient.SSHConfig{
		Host:    getDomainFromPath(pth),
		User:    "kl",
		KeyPath: path.Join(xdg.Home, ".ssh", "id_rsa"),
		SSHPort: localEnv.SSHPort,
	}); err != nil {
		return err
	}

	// command.Stderr = os.Stderr
	// command.Stdin = os.Stdin
	// command.Stdout = os.Stdout
	// if err := command.Run(); err != nil {
	// 	return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err.Error())
	// }
	return nil
}
