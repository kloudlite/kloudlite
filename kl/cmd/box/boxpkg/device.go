package boxpkg

import (
	cl "github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
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
		functions.Warn("starting vpn")
		if err := p.Start(); err != nil {
			return err
		}
	}

	return nil
}
