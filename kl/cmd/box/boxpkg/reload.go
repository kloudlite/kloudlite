package boxpkg

import (
	"encoding/json"
	"os"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (c *client) Reload() error {
	defer c.spinner.Start("Reloading environments please wait")()

	envs, mmap, err := server.GetLoadMaps()
	if err != nil {
		return err
	}

	// local setup
	kConf, err := c.loadConfig(mmap, envs)
	if err != nil {
		return err
	}

	conf, err := json.Marshal(kConf)
	if err != nil {
		return err
	}

	if err := os.WriteFile("/tmp/sample.json", conf, os.ModePerm); err != nil {
		return err
	}

	fn.Warn("configuration changes have been applied. To ensure these changes take effect, please restart your SSH/IDE sessions.")

	return nil
}
