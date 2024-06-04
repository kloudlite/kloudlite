package boxpkg

import (
	"fmt"

	cl "github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
)

func (c *client) Reload() error {
	defer spinner.Client.Start("Reloading environments please wait")()

	if err := server.SyncDevboxJsonFile(); err != nil {
		return err
	}

	fn.Warn("configuration changes have been applied. To ensure these changes take effect, please restart your SSH/IDE sessions.")

	return cl.ExecPackageCommand(fmt.Sprintf("devbox install%s", func() string {
		if c.verbose {
			return ""
		}
		return " -q"
	}()))
}
