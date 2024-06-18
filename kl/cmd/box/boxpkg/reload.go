package boxpkg

import (
	"github.com/kloudlite/kl/domain/server"

	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func (c *client) Reload() error {
	defer spinner.Client.Start("Reloading environments please wait")()

	if err := server.SyncBoxHash(); err != nil {
		return err
	}
	return nil
}
