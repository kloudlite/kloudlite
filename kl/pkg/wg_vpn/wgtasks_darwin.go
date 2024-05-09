package wg_vpn

import (
	"errors"
	"fmt"
	"github.com/kloudlite/kl/constants"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
	"github.com/kloudlite/kl/pkg/functions"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

const (
	ifName string = "utun2464"
)

func ipRouteAdd(ip string, interfaceIp string, deviceName string, verbose bool) error {
	return ExecCmd(fmt.Sprintf("route -n add -net %s %s", ip, interfaceIp), verbose)
}

//func getNetworkServices(verbose bool) ([]string, error) {
//	output, err := exec.Command("networksetup", "-listallnetworkservices").Output()
//	if err != nil {
//		functions.Log(string(output))
//		return nil, err
//	}
//	lines := strings.Split(string(output), "\n")
//	var networkServices []string
//	for _, line := range lines[1:] {
//		if line == "" {
//			continue
//		}
//		networkServices = append(networkServices, line)
//	}
//	return networkServices, err
//}

func getCurrentDns(verbose bool) ([]string, error) {
	networkService := "Wi-Fi"
	if verbose {
		functions.Log(fmt.Sprintf("[#] networksetup -getdnsservers %s", networkService))
	}
	output, err := exec.Command("networksetup", "-getdnsservers", networkService).Output()
	if err != nil {
		functions.Log(fmt.Sprintf("[#] %s", string(output)))
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var dnsServers []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		dnsServers = append(dnsServers, line)
	}
	return dnsServers, nil
}

func SetDeviceIp(deviceIp net.IPNet, _ string, verbose bool) error {
	return ExecCmd(fmt.Sprintf("ifconfig %s %s %s", ifName, deviceIp.IP.String(), deviceIp.IP.String()), verbose)
}

func StartService(_ string, verbose bool) error {

	t, err := tun.CreateTUN(ifName, device.DefaultMTU)
	if err != nil {
		return err
	}

	fileUAPI, err := ipc.UAPIOpen(ifName)
	if err != nil {
		return err
	}
	var logger *device.Logger
	if verbose {
		logger = device.NewLogger(
			device.LogLevelSilent,
			fmt.Sprintf("[%s]", ifName),
		)
	} else {
		logger = device.NewLogger(
			device.LogLevelVerbose,
			fmt.Sprintf("[%s]", ifName),
		)
	}

	d := device.NewDevice(t, conn.NewDefaultBind(), logger)
	logger.Verbosef("Device started")
	errs := make(chan error)
	term := make(chan os.Signal, 1)
	uapi, err := ipc.UAPIListen(ifName, fileUAPI)
	if err != nil {
		logger.Errorf("Failed to listen on uapi socket: %v", err)
		os.Exit(1)
	}
	go func() {
		for {
			c, err := uapi.Accept()
			if err != nil {
				errs <- err
				return
			}
			go d.IpcHandle(c)
		}
	}()

	logger.Verbosef("UAPI listener started")
	signal.Notify(term, syscall.SIGTERM)
	signal.Notify(term, syscall.SIGKILL)
	signal.Notify(term, os.Interrupt)

	select {
	case <-term:
	case <-errs:
	case <-d.Wait():
	}
	_ = uapi.Close()
	d.Close()
	logger.Verbosef("Shutting down")
	return nil
}

func StopService(verbose bool) error {
	output, err := exec.Command("pgrep", "-f", fmt.Sprintf("%s %s", flags.CliName, "vpn start-fg")).Output()
	if err != nil {
		return err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return err
	}
	p, err := os.FindProcess(int(i))
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("process not found")
	}

	err = syscall.Kill(int(i), syscall.SIGTERM)
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		dnsServers, err := getCurrentDns(verbose)
		if err != nil {
			return err
		}

		servers, err := client.GetDns()
		if err != nil {
			return err
		}

		filteredDnsServers := []string{}

		for _, dnsServer := range dnsServers {
			if !slices.Contains(servers, dnsServer) {
				filteredDnsServers = append(filteredDnsServers, dnsServer)
			}
		}

		if err := setDnsServers(func() []net.IP {
			var ipNets []net.IP
			for _, dnsServer := range filteredDnsServers {
				ipNets = append(ipNets, net.ParseIP(dnsServer))
			}
			return ipNets
		}(), "Wi-Fi", verbose); err != nil {
			return err
		}

	}

	return nil
}

//func setDnsServer(dnsServer net.IP, d string, verbose bool) error {
//	return ExecCmd(fmt.Sprintf("networksetup -setdnsservers %s %s", d, dnsServer.String()), verbose)
//}

func setDnsSearchDomain(networkService string, localSearchDomains []string) error {
	if localSearchDomains == nil {
		return ExecCmd(fmt.Sprintf("networksetup -setsearchdomains %s %s", networkService, "Empty"), false)
	}
	return ExecCmd(fmt.Sprintf("networksetup -setsearchdomains %s %s", networkService, strings.Join(localSearchDomains, " ")), false)
}

func getDnsSearchDomain(networkService string) ([]string, error) {
	d, err := exec.Command("networksetup", "-getsearchdomains", networkService).Output()
	if err != nil {
		return nil, err
	}
	domains := strings.Split(strings.TrimSpace(string(d)), "\n")
	if domains[0] == constants.NoExistingSearchDomainError {
		return domains, errors.New("no existing search domain found")
	}
	return domains, nil
}

func SetDnsSearch() error {
	data, err := client.GetExtraData()
	if err != nil {
		return err
	}
	currentDns, err := getCurrentDns(false)
	if err != nil {
		return err
	}

	if data.DnsAdded {
		uniqueDns := make(map[string]bool)
		for _, dns := range data.DnsValues {
			uniqueDns[dns] = true
		}
		for _, val := range currentDns {
			uniqueDns[val] = true
		}
		var combinedDns []string
		for val := range uniqueDns {
			combinedDns = append(combinedDns, val)
		}
		data.DnsValues = combinedDns
	} else {
		data.DnsValues = currentDns
		data.DnsAdded = true
	}
	if err := client.SaveExtraData(data); err != nil {
		return err
	}

	var privateIPs []net.IP
	for _, ipStr := range data.DnsValues {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			functions.Log("Invalid IP address: %s\n", ipStr)
			continue
		}
		if isPrivateIP(ip) {
			privateIPs = append(privateIPs, net.ParseIP(ipStr))
		}
	}

	err = setDnsServers(privateIPs, constants.NetworkService, false)
	if err != nil {
		return err
	}

	searchDomains, err := getDnsSearchDomain(constants.NetworkService)
	if err == nil {
		if slices.Contains(searchDomains, constants.LocalSearchDomains) {
			return nil
		}
		searchDomains = append(searchDomains, constants.LocalSearchDomains)
		err1 := setDnsSearchDomain(constants.NetworkService, searchDomains)
		if err1 != nil {
			return nil
		}
	} else {
		searchDomains[0] = constants.LocalSearchDomains
		err1 := setDnsSearchDomain(constants.NetworkService, searchDomains)
		if err1 != nil {
			return nil
		}
	}
	data.SearchDomainAdded = true
	if err := client.SaveExtraData(data); err != nil {
		return err
	}
	return nil
}

func UnsetDnsSearch() error {
	data, err := client.GetExtraData()
	if err != nil {
		return err
	}

	if data.DnsAdded {
		ips := make([]net.IP, 0)
		for _, dns := range data.DnsValues {
			ips = append(ips, net.ParseIP(dns))
		}
		if err := setDnsServers(ips, constants.NetworkService, false); err != nil {
			return err
		}
	}
	data.DnsAdded = false
	data.DnsValues = nil

	if err := client.SaveExtraData(data); err != nil {
		return err
	}

	if data.SearchDomainAdded {
		searchDomains, err := getDnsSearchDomain(constants.NetworkService)
		if err != nil {
			return err
		}
		searchDomains = functions.RemoveFromArray(constants.LocalSearchDomains, searchDomains)
		if err = setDnsSearchDomain(constants.NetworkService, searchDomains); err != nil {
			return err
		}
		data.SearchDomainAdded = false
		if err := client.SaveExtraData(data); err != nil {
			return err
		}
	}
	return nil
}

func isPrivateIP(ip net.IP) bool {
	private := false
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			private = true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			private = true
		case ip4[0] == 192 && ip4[1] == 168:
			private = true
		}
	}
	return private
}

func setDnsServers(dnsServers []net.IP, _ string, verbose bool) error {

	if len(dnsServers) == 0 {
		if verbose {
			functions.Warn("dns not configured, using old dns server")
		}
		return nil
	}

	return ExecCmd(fmt.Sprintf("networksetup -setdnsservers %s %s", constants.NetworkService, func() string {
		var dns []string
		for _, v := range dnsServers {
			dns = append(dns, v.To4().String())
		}
		return strings.Join(dns, " ")
	}()), verbose)
}
