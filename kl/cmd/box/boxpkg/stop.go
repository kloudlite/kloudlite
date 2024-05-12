package boxpkg

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Stop() error {
	defer c.spinner.Start("stopping container please wait")()

	if err := c.StopVpn(); err != nil {
		fn.Warn("[#] vpn stop error:", err.Error())
	}

	cr, err := c.getContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr {
		return fmt.Errorf("no running container found")
	}

	crPath := cr.Labels[CONT_PATH_KEY]

	if c.verbose {
		fn.Logf("stopping container of: %s", text.Blue(crPath))
	}

	if err := c.cli.ContainerStop(c.Context(), cr.Name, container.StopOptions{}); err != nil {
		return fmt.Errorf("error stoping container: %s", err)
	}

	if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	if c.verbose {
		fn.Logf("stopped container of: %s", text.Blue(crPath))
	}
	return nil
}
