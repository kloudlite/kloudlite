package boxpkg

import (
	cl "github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/kloudlite/kl/domain/server"
)

func (c *client) ensureVpnConnected() error {
	if err := cl.EnsureAppRunning(); err != nil {
		return err
	}

	p, err := proxy.NewProxy(c.verbose, false)
	if err != nil {
		return err
	}

	if !server.CheckDeviceStatus() {
		if err := p.Start(); err != nil {
			return err
		}
	}

	return nil
}
