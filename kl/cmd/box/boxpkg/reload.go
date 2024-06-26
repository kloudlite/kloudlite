package boxpkg

import (
	"os"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
)

func (c *client) Reload() error {

	wpath, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := hashctrl.SyncBoxHash(wpath); err != nil {
		return err
	}
	return nil
}
