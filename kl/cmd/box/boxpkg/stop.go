package boxpkg

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	cl "github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/sshclient"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Stop() error {
	defer spinner.Client.Start("stopping container please wait")()

	cr, err := c.GetContainer(map[string]string{
		CONT_MARK_KEY: "true",
		CONT_NAME_KEY: c.containerName,
	})
	if err != nil && err != NotFoundErr {
		return err
	}

	if err == NotFoundErr {

		conts, err2 := c.ListAllBoxes()
		if err != nil && err != NotFoundErr {
			return err2
		}

		if err2 == nil {
			c.PrintBoxes(conts)
			return fmt.Errorf("no running container found in current workspace")
		}

		return fmt.Errorf("no running container found in current workspace")
	}

	crPath := cr.Labels[CONT_PATH_KEY]

	if c.verbose {
		fn.Logf("stopping container of: %s", text.Blue(crPath))
	}

	if cr.State != ContStateExited && cr.State != ContStateCreated {
		if err := c.cli.ContainerKill(c.Context(), cr.Name, "SIGKILL"); err != nil {
			return fmt.Errorf("error stoping container: %s", err)
		}
	}

	if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	localEnv, err := cl.EnvOfPath(crPath)
	if err != nil {
		return err
	}

	if localEnv.SSHPort != 0 {
		p, err := proxy.NewProxy(false)
		if err != nil {
			return err
		}

		if _, err := p.RemoveAllFwd(sshclient.StartCh{
			SshPort: fmt.Sprint(localEnv.SSHPort),
		}); err != nil {
			return err
		}
	}

	if c.verbose {
		fn.Logf("stopped container of: %s", text.Blue(crPath))
	}
	return nil
}

func (c *client) StopAll() error {
	defer spinner.Client.Start("stopping container please wait")()

	crs, err := c.listContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})

	if err != nil && err != NotFoundErr {
		return err
	}

	if err == NotFoundErr {
		fn.Warn("no running containers found in any workspace")
	}

	for _, cr := range crs {
		crPath := cr.Labels[CONT_PATH_KEY]

		if c.verbose {
			fn.Logf("stopping container of: %s", text.Blue(crPath))
		}

		if cr.State != ContStateExited && cr.State != ContStateCreated {
			if err := c.cli.ContainerKill(c.Context(), cr.Name, "SIGKILL"); err != nil {
				fn.Warnf("error stoping container: %s", err.Error())
				continue
			}
		}

		if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
			fn.Warnf("failed to remove container: %s", err.Error())
			continue
		}

		if c.verbose {
			fn.Logf("stopped container of: %s", text.Blue(crPath))
		}
	}

	return nil
}

func (c *client) StopCont(cr *Cntr) error {
	defer spinner.Client.Start("stopping container please wait")()

	crPath := cr.Labels[CONT_PATH_KEY]

	if c.verbose {
		fn.Logf("stopping container of: %s", text.Blue(crPath))
	}

	if cr.State != ContStateExited && cr.State != ContStateCreated {
		if err := c.cli.ContainerKill(c.Context(), cr.Name, "SIGKILL"); err != nil {
			return fmt.Errorf("error stoping container: %s", err)
		}
	}

	if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	localEnv, err := cl.EnvOfPath(crPath)
	if err != nil {
		return err
	}

	if localEnv.SSHPort != 0 {
		p, err := proxy.NewProxy(false)
		if err != nil {
			return err
		}

		if _, err := p.RemoveAllFwd(sshclient.StartCh{
			SshPort: fmt.Sprint(localEnv.SSHPort),
		}); err != nil {
			return err
		}
	}

	if c.verbose {
		fn.Logf("stopped container of: %s", text.Blue(crPath))
	}
	return nil
}
