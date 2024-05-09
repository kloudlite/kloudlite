package wg_vpn

import (
	"fmt"
	"net"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func ResetDnsServers(devName string, verbose bool) error {

	bkDns, err := client.BackupDns()
	if err != nil {
		fn.PrintError(fmt.Errorf("failed to get active dns servers: %w", err))
		return nil
	}

	if len(bkDns) == 0 {
		return nil
	}

	ips := make([]net.IP, 0)
	for _, v := range bkDns {
		ips = append(ips, net.ParseIP(v))
	}

	if err := setDnsServers(ips, devName, verbose); err != nil {
		return err
	}

	if err := client.SetBackupDns([]string{}); err != nil {
		fn.PrintError(err)
	}

	return nil
}

func SetDnsServers(dnsServers []net.IP, devName string, verbose bool) error {

	warn := func(str ...interface{}) {
		if verbose {
			fn.Warn(str)
		}
	}

	log := func(str ...interface{}) {
		if verbose {
			fn.Warn(str)
		}
	}

	if len(dnsServers) == 0 {
		warn("# dns server is not configured")
		return nil
	}

	// backup ip
	if err := func() error {
		currDns, _ := getCurrentDns(verbose)
		if len(currDns) == 0 {
			warn("# no dns server is configured to backup")
			return nil
		}

		bkDns, err := client.GetDns()
		if err != nil {
			return err
		}

		if len(bkDns) != 0 {
			return nil
		}

		for _, i := range currDns {
			found := false
			for _, j := range dnsServers {
				if j.To4().String() == i {
					found = true
					break
				}
			}
			if !found {
				dnsServers = append(dnsServers, net.ParseIP(i))
			}
		}

		return client.SetBackupDns(currDns)
	}(); err != nil {
		return err
	}

	if verbose {
		log("# updating dns server")
	}

	return setDnsServers(dnsServers, devName, verbose)
}
