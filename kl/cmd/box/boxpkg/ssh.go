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
	defer c.spinner.Stop()

	cont, err := c.getContainer()
	if err != nil {
		return err
	}

	if cont.Name == "" {
		err := c.Start()
		if err == nil {

			c.spinner.Start("waiting for container to be ready")
			time.Sleep(5 * time.Second)
			c.spinner.Stop()
		}

		if err != nil && err != containerNotStartedErr {
			return err
		}
	}

	if cont.Name != "" && c.containerName != cont.Name {
		// fn.Warnf("\ncontainer already running, using container of workspace '%s'", c.cwd)

		// if needed to restart server for unique workspace then uncomment below line

		err := c.Start()
		if err == nil {
			c.spinner.Start("waiting for container to be ready")
			time.Sleep(5 * time.Second)
			c.spinner.Stop()
		}

		if err != nil && err != containerNotStartedErr {
			return err
		}
	}

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
