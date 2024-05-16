package vpn

import (
	"fmt"

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"
)

func connect(verbose bool, options ...fn.Option) error {

	_ = client.SetLoading(true)

	success := false
	defer func() {
		if !success {
			_ = wg_vpn.StopService(verbose)
		}

		_ = client.SetLoading(false)
	}()

	switch flags.CliName {
	case constants.CoreCliName:
		_, err := server.EnsureEnv(nil, options...)
		if err != nil {
			return err
		}
	case constants.InfraCliName:
		_, err := server.EnsureAccount()
		if err != nil {
			return err
		}
	}

	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	// TODO: handle this error later
	if err = wg_vpn.StartServiceInBg(ifName, configFolder); err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	}

	// if err := ensureAppRunning(); err != nil {
	// 	fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	// }

	//if err = wg_vpn.SetDnsSearch(); err != nil {
	//	return err
	//}

	if err := startConfiguration(verbose, options...); err != nil {
		_ = wg_vpn.ResetDnsServers(ifName, verbose)
		return err
	}
	success = true

	data, err := client.GetExtraData()
	if err != nil {
		return err
	}
	data.VpnConnected = true
	if err := client.SaveExtraData(data); err != nil {
		return err
	}
	return nil
}

func disconnect(verbose bool) error {

	// if err := ensureAppRunning(); err != nil {
	// 	fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	// }

	if err := wg_vpn.StopService(verbose); err != nil {
		return err
	}

	data, err := client.GetExtraData()
	if err != nil {
		return err
	}
	data.VpnConnected = false
	if err := client.SaveExtraData(data); err != nil {
		return err
	}

	if err := wg_vpn.ResetDnsServers(ifName, verbose); err != nil {
		return err
	}
	// if err = wg_vpn.UnsetDnsSearch(); err != nil {
	// 	return err
	// }
	return nil
}
