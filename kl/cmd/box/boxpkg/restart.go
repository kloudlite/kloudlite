package boxpkg

import (
	"github.com/kloudlite/kl/domain/envclient"
)

func (c *client) Restart() error {
	if envclient.InsideBox() {
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
