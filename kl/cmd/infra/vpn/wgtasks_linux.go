package vpn

import (
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

func connect(verbose bool) error {
	success := false
	defer func() {
		if !success {
			wg_vpn.StopService(verbose)
		}
	}()

	devName, err := server.EnsureDevice()
	if err != nil {
		return err
	}

	// TODO: handle this error later
	_ = wg_vpn.StartService(devName, verbose)

	if err := startConfiguration(verbose); err != nil {
		return err
	}
	success = true
	return nil
}

func disconnect(verbose bool) error {
	return wg_vpn.StopService(verbose)
}
