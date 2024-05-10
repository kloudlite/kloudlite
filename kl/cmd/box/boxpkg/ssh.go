package boxpkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/adrg/xdg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Ssh() error {
	cont, err := c.getContainer()
	if err != nil {
		return err
	}

	s := spinner.NewSpinner("waiting for container to be ready")

	if cont.Name == "" {
		if err := c.Start(); err == nil {
			s.Start()
			time.Sleep(5 * time.Second)
			s.Stop()

		}

		if err != nil && err != containerNotStartedErr {
			return err
		}
	}

	if cont.Name != "" && c.containerName != cont.Name {
		fn.Warnf("\ncontainer already running, using container of workspace '%s'", c.cwd)
	}

	command := exec.Command("ssh", "kl@localhost", "-p", CONTAINER_PORT, "-i", path.Join(xdg.Home, ".ssh", "id_rsa"))

	fn.Logf("\n%s: %s\n", text.Bold("ssh command"), text.Blue(command.String()))

	if c.verbose {
		fn.Log(command.String())
	}

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err)
	}
	return nil

}
