package boxpkg

import (
	"fmt"
	"os"
	"os/exec"

	fn "github.com/kloudlite/kl/pkg/functions"
)

func (c *client) Exec() error {
	cr, err := c.getContainer()
	if err != nil {
		return err
	}

	if cr.Name == "" {
		return fmt.Errorf("no running container found")
	}

	command := exec.Command("docker", "exec", "-it", cr.Name, "bash")

	if c.verbose {
		fn.Log(command.String())
	}

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		fn.PrintError(fmt.Errorf("failed to run command: %s", err))
		return err
	}
	return nil

}
