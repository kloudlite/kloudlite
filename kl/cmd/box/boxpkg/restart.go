package boxpkg

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Restart() error {

	if envclient.InsideBox() {
		fn.Logf(text.Yellow("[#] this will restart current workspace and action will terminate all the processes running in the container. do you want to proceed? [Y/n] "))
		if !fn.Confirm("y", "y") {
			return fn.NewE(fn.NewE(UserCanceled))
		}
		path, err := envclient.GetWorkspacePath()
		if err != nil {
			return err
		}
		return c.restartContainer(path)
	}
	return c.restartBoxContainer()
}

func (c *client) restartBoxContainer() error {
	crlist, err := c.cli.ContainerList(c.cmd.Context(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", CONT_MARK_KEY, "true")),
		),
	})
	if err != nil {
		return fn.NewE(err)
	}

	if len(crlist) != 0 {
		fn.Logf(text.Yellow("[#] this will restart current workspace and action will terminate all the processes running in the container. do you want to proceed? [Y/n]"))
		if !fn.Confirm("y", "y") {
			return fn.NewE(fn.NewE(UserCanceled))
		}
	}

	timeOut := 0
	for _, cr := range crlist {
		if cr.Labels["kl-k3s"] == "true" {
			continue
		}
		if err := c.cli.ContainerStop(c.cmd.Context(), cr.ID, container.StopOptions{
			Signal:  "SIGKILL",
			Timeout: &timeOut,
		}); err != nil {
			return err
		}
		if err := c.cli.ContainerRemove(c.cmd.Context(), cr.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}

	if err = c.Start(); err != nil {
		return err
	}
	return nil
}
