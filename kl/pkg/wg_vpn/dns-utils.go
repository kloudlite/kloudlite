package wg_vpn

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
)

func setLinuxDnsServers(dnsServers []net.IP, verbose bool) error {
	ed, err := client.ActiveDns()
	if err != nil {
		return err
	}

	if len(dnsServers) == 0 {
		functions.Warn("# dns server is not configured")
		return nil
	}

	// backup ip
	if err := func() error {
		s, err := getCurrentDns(verbose)
		if err != nil {
			functions.PrintError(err)
			return nil
		}

		if len(ed) != 0 {
			functions.Warn("# dns server is not configured")
			return nil
		}

		if len(s) == 0 {
			functions.Warn("# dns server is not configured")
			return nil
		}

		return nil
	}(); err != nil {
		return err
	}

	if verbose {
		functions.Log("# updating dns server")
	}

	ips := []string{}
	for _, v := range dnsServers {
		ips = append(ips, fmt.Sprintf("nameserver %s", v.To4().String()))
	}

	if err := os.WriteFile("/etc/resolv.conf", []byte(strings.Join(ips, "\n")), 0644); err != nil {
		return err
	}

	return nil
}

func setDnsServers(dnsServers []net.IPNet, inf string, verbose bool) error {
	return ExecCmd(fmt.Sprintf("networksetup -setdnsservers %s %s", inf, func() string {
		var dns []string
		for _, v := range dnsServers {
			dns = append(dns, v.IP.String())
		}
		return strings.Join(dns, " ")
	}()), verbose)
}
