package boxpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/adrg/xdg"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (c *client) Ssh() error {
	defer c.spinner.Start("preparing to ssh")()

	cr, err := c.getContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr || (err == nil && c.containerName != cr.Name) {
		err := c.Start()
		if err == nil {
			time.Sleep(5 * time.Second)
		}

		if err != nil && err != containerNotStartedErr {
			return err
		}
	}

	c.spinner.Stop()
	command := exec.Command("ssh", "kl@localhost", "-p", CONTAINER_PORT, "-i", path.Join(xdg.Home, ".ssh", "id_rsa"), "-oStrictHostKeyChecking=no")

	if c.verbose {
		fn.Log(command.String())
	}

	// command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err)
	}
	return nil

}
