package boxpkg

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Stop() error {
	defer c.spinner.Start("stopping container please wait")()

	cr, err := c.getContainer(map[string]string{
		CONT_MARK_KEY: "true",
		CONT_NAME_KEY: c.containerName,
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr {
		c.ListBox()
		return fmt.Errorf("no running container found in current workspace")
	}

	crPath := cr.Labels[CONT_PATH_KEY]

	if c.verbose {
		fn.Logf("stopping container of: %s", text.Blue(crPath))
	}

	if cr.State != ContStateExited {
		if err := c.cli.ContainerKill(c.Context(), cr.Name, "SIGKILL"); err != nil {
			return fmt.Errorf("error stoping container: %s", err)
		}
	}

	if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	if c.verbose {
		fn.Logf("stopped container of: %s", text.Blue(crPath))
	}
	return nil
}

func (c *client) StopAll() error {
	defer c.spinner.Start("stopping container please wait")()

	crs, err := c.listContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr {
		c.ListBox()
		return fmt.Errorf("no running container found in current workspace")
	}

	for _, cr := range crs {
		crPath := cr.Labels[CONT_PATH_KEY]

		if c.verbose {
			fn.Logf("stopping container of: %s", text.Blue(crPath))
		}

		if cr.State != ContStateExited {
			if err := c.cli.ContainerKill(c.Context(), cr.Name, "SIGKILL"); err != nil {
				fn.Warnf("error stoping container: %s", err)
				continue
			}
		}

		if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
			fn.Warnf("failed to remove container: %s", err)
			continue
		}

		if c.verbose {
			fn.Logf("stopped container of: %s", text.Blue(crPath))
		}
	}

	return nil
}
