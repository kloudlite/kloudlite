package vpn

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/wg_vpn"
	"github.com/miekg/dns"
)

func getCurrentDns() ([]string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, err
	}

	return config.Servers, nil
}

func connect(verbose bool) error {
	success := false
	defer func() {
		if !success {
			_ = wg_vpn.StopService(verbose)
		}
	}()

	configFolder, err := client.GetConfigFolder()
	if err != nil {
		return err
	}

	if err := wg_vpn.StartServiceInBg(ifName, configFolder); err != nil {
		return err
	}

	if err := startConfiguration(connectVerbose); err != nil {
		return err
	}

	success = true
	return nil
}

func disconnect(verbose bool) error {
	return wg_vpn.StopService(verbose)
}
