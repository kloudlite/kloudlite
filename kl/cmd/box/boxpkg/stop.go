package boxpkg

import (
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func (c *client) Stop() error {
	return c.stopContainer(c.cwd)
}

func (c *client) StopAll() error {
	defer spinner.Client.UpdateMessage("stopping container please wait")()

	crs, err := c.listContainer(map[string]string{
		CONT_MARK_KEY: "true",
	})

	if err != nil && err != NotFoundErr {
		return functions.NewE(err)
	}

	if err == NotFoundErr {
		fn.Warn("no running containers found in any workspace")
	}

	for _, cr := range crs {
		if err := c.stopContainer(cr.Labels[CONT_PATH_KEY]); err != nil {
			return functions.NewE(err)
		}
	}

	return nil
}
