package vpn

import (
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/wg_vpn"

	"github.com/kloudlite/kl/domain/client"
)

func connect(verbose bool, options ...fn.Option) error {
	success := false
	defer func() {
		if !success {
			wg_vpn.StopService(verbose)
		}
	}()

	switch flags.CliName {
	case constants.CoreCliName:
		_, err := server.EnsureProject()
		if err != nil {
			return err
		}
	}

	devName, err := client.CurrentDeviceName()
	if err != nil {
		return err
	}

	// TODO: handle this error later
	_ = wg_vpn.StartService(devName, verbose)

	if err := startConfiguration(verbose, options...); err != nil {
		return err
	}
	success = true
	return nil
}

func disconnect(verbose bool) error {
	return wg_vpn.StopService(verbose)
}
