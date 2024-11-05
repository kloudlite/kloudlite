package boxpkg

import (
	"context"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Reload() error {

	wpath, err := os.Getwd()
	if err != nil {
		return fn.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(c.apic, c.fc, wpath); err != nil {
		return fn.NewE(err)
	}
	return nil
}

func (c *client) ConfirmBoxRestart() error {
	if envclient.InsideBox() {
		return nil
	}
	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			dockerLabelFilter(CONT_MARK_KEY, "true"),
			dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
			dockerLabelFilter(CONT_PATH_KEY, c.cwd),
		),
	})

	if err != nil {
		return err
	}
	if len(existingContainers) == 0 {
		return nil
	}

	pe, err := hashctrl.BoxHashFile(c.cwd)
	if err != nil {
		return err
	}

	cntrHash := existingContainers[0].Labels[KLCONFIG_HASH_KEY]

	if pe.KLConfHash == cntrHash {
		return nil
	}

	fn.Logf(text.Yellow("[#] Environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
	if !fn.Confirm("Y", "Y") {
		return nil
	}

	if err = c.Stop(); err != nil {
		return err
	}

	if err = c.Start(); err != nil {
		return err
	}

	return nil
}
