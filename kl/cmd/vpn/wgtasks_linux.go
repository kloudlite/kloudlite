package vpn

import (
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/wg_vpn"

	"github.com/kloudlite/kl/domain/client"
)

func connect(verbose bool) error {
	success := false
	defer func() {
		if !success {
			wg_vpn.StopService(verbose)
		}
	}()

	_, err := server.EnsureProject()
	if err != nil {
		return err
	}

	devName, err := client.CurrentDeviceName()
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
