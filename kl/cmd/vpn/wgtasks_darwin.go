package vpn

import (
	fn "github.com/kloudlite/kl/pkg/functions"
	wg_svc "github.com/kloudlite/kl/pkg/wg_vpn/wg_service"
)

func connect(verbose bool, options ...fn.Option) error {

	if err := wg_svc.EnsureInstalled(); err != nil {
		return err
	}

	if err := wg_svc.EnsureAppRunning(); err != nil {
		return err
	}

	success := false
	defer func() {
		if !success {
			_ = wg_svc.StopVpn(verbose)
		}
	}()

	if err := startConfiguration(connectVerbose, options...); err != nil {
		return err
	}

	success = true
	return nil
}

func disconnect(verbose bool) error {
	if err := wg_svc.EnsureInstalled(); err != nil {
		return err
	}

	if err := wg_svc.EnsureAppRunning(); err != nil {
		return err
	}

	return wg_svc.StopVpn(verbose)
}
