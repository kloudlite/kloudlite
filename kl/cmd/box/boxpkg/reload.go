package boxpkg

import (
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/pkg/functions"
)

func (c *client) Reload() error {

	wpath, err := os.Getwd()
	if err != nil {
		return functions.NewE(err)
	}

	if err := hashctrl.SyncBoxHash(wpath); err != nil {
		return functions.NewE(err)
	}
	return nil
}
