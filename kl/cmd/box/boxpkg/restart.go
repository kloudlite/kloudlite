package boxpkg

import (
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
	if err := c.Stop(); err != nil {
		return err
	}

	if err := c.Start(); err != nil {
		return err
	}
	return nil
}
