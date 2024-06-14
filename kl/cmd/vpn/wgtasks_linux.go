package vpn

import (
	"fmt"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn"

	"github.com/kloudlite/kl/domain/client"
)

func connect(verbose bool, options ...fn.Option) error {

	client.SetLoading(true)

	success := false
	defer func() {
		if !success {
			_ = wg_vpn.StopService(verbose)

			if !wg_vpn.IsSystemdReslov() {
				wg_vpn.ResetDnsServers(ifName, verbose)
			}
		}

		client.SetLoading(false)

	}()

	// TODO: handle this error later
	if err := wg_vpn.StartService(ifName, verbose); err != nil {
		fn.Log(text.Yellow(fmt.Sprintf("[#] %s", err)))
	}

	if err := startConfiguration(verbose, options...); err != nil {
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

	if !wg_vpn.IsSystemdReslov() {
		if err := wg_vpn.ResetDnsServers(ifName, verbose); err != nil {
			return err
		}
	}

	dc, err := client.GetDeviceContext()
	if err != nil {
		return err
	}

	dc.DeviceDns = []string{}

	if err := client.WriteDeviceContext(dc); err != nil {
		return err
	}
	return nil
}
