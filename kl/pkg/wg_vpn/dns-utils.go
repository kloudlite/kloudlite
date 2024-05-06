package wg_vpn

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func setLinuxDnsServers(dnsServers []net.IP, verbose bool) error {
	if len(dnsServers) == 0 {
		fn.Warn("# dns server is not configured")
		return nil
	}

	// backup ip
	if err := func() error {
		s, _ := getCurrentDns(verbose)

		if len(s) == 0 || dnsServers[0].To4().String() != s[0] {
			if err := client.SetActiveDns([]string{
				s[0],
			}); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return err
	}

	if verbose {
		fn.Log("# updating dns server")
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

func ResetLinuxDnsServers() error {
	s, err := client.ActiveDns()
	if err != nil {
		fn.PrintError(fmt.Errorf("failed to get active dns servers: %w", err))
		return nil
	}

	if len(s) == 0 {
		return nil
	}

	ips := []string{}
	for _, v := range s {
		ips = append(ips, fmt.Sprintf("nameserver %s", v))
	}

	if err := os.WriteFile("/etc/resolv.conf", []byte(strings.Join(ips, "\n")), 0644); err != nil {
		return err
	}

	if err := client.SetActiveDns([]string{}); err != nil {
		fn.PrintError(err)
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
