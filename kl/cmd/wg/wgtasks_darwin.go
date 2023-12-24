package wg

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
	"k8s.io/utils/strings/slices"
)

const (
	KlWgInterface = "utun1729"
)

func connect(verbose bool) error {
	success := false
	defer func() {
		if !success {
			_ = stopService(verbose)
		}
	}()

	startServiceInBg()
	if err := startConfiguration(connectVerbose); err != nil {
		return err
	}

	success = true
	return nil
}
func disconnect(verbose bool) error {
	return nil
}

func ipRouteAdd(ip string, interfaceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("route -n add -net %s %s", ip, interfaceIp), verbose)
}

func getNetworkServices(verbose bool) ([]string, error) {
	output, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		common.Log(string(output))
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var networkServices []string
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		networkServices = append(networkServices, line)
	}
	return networkServices, err
}

func getDNSServers(networkService string, verbose bool) ([]string, error) {
	if verbose {
		common.Log(fmt.Sprintf("[#] networksetup -getdnsservers %s", networkService))
	}
	output, err := exec.Command("networksetup", "-getdnsservers", networkService).Output()
	if err != nil {
		common.Log(fmt.Sprintf("[#] %s", string(output)))
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(output), "\n")
	var dnsServers []string
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		dnsServers = append(dnsServers, line)
	}
	return dnsServers, nil
}

func removeDNSServer(networkService string, dnsServer string, verbose bool) error {
	dnsServers, err := getDNSServers(networkService, verbose)
	if err != nil {
		return err
	}
	newDnsServers := slices.Filter([]string{}, dnsServers, func(d string) bool {
		return d != dnsServer
	})
	if len(newDnsServers) == 0 {
		_ = execCmd(fmt.Sprintf("networksetup -setdnsservers %q empty", networkService), verbose)
	} else {
		_ = execCmd(fmt.Sprintf("networksetup -setdnsservers %q %s", networkService, strings.Join(newDnsServers, " ")), verbose)
	}
	return err
}

func addDNSServer(networkService string, dnsServer string, verbose bool) error {
	dnsServers, err := getDNSServers(networkService, verbose)
	if err != nil {
		return err
	}
	newDnsServers := append(dnsServers, dnsServer)
	return execCmd(fmt.Sprintf("networksetup -setdnsservers %q %s 8.8.8.8", networkService, strings.Join(newDnsServers, " ")), verbose)
}

func setDNS(dns []net.IP, verbose bool) error {
	err := server.SetActiveDns(func() []string {
		var dnsServers []string
		for _, d := range dns {
			dnsServers = append(dnsServers, d.String())
		}
		return dnsServers
	}())
	if err != nil {
		return err
	}
	services, err := getNetworkServices(verbose)
	if err != nil {
		return err
	}
	for _, service := range services {
		for _, ip := range dns {
			err = addDNSServer(service, ip.String(), verbose)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func resetDNS(verbose bool) error {
	dns, err := server.ActiveDns()
	if err != nil {
		return err
	}
	services, err := getNetworkServices(verbose)
	if err != nil {
		return err
	}
	for _, d := range dns {
		for _, service := range services {
			err = removeDNSServer(service, d, verbose)
			if err != nil {
				return err
			}
		}
	}
	_ = server.SetActiveDns([]string{})
	return nil
}

func setDeviceIp(deviceIp string, verbose bool) error {
	return execCmd(fmt.Sprintf("ifconfig %s %s %s", KlWgInterface, deviceIp, deviceIp), verbose)
}
func startService(verbose bool) error {
	t, err := tun.CreateTUN(KlWgInterface, device.DefaultMTU)
	if err != nil {
		return err
	}
	fileUAPI, err := ipc.UAPIOpen(KlWgInterface)
	if err != nil {
		return err
	}
	var logger *device.Logger
	if verbose {
		logger = device.NewLogger(
			device.LogLevelSilent,
			fmt.Sprintf("[%s]", KlWgInterface),
		)
	} else {
		logger = device.NewLogger(
			device.LogLevelVerbose,
			fmt.Sprintf("[%s]", KlWgInterface),
		)
	}

	d := device.NewDevice(t, conn.NewDefaultBind(), logger)
	logger.Verbosef("Device started")
	errs := make(chan error)
	term := make(chan os.Signal, 1)
	uapi, err := ipc.UAPIListen(KlWgInterface, fileUAPI)
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

func stopService(verbose bool) error {
	output, err := exec.Command("pgrep", "-f", "kl wg connect --foreground").Output()
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
	err = resetDNS(verbose)
	if err != nil {
		return err
	}
	err = syscall.Kill(int(i), syscall.SIGTERM)
	if err != nil {
		return err
	}
	return nil
}
