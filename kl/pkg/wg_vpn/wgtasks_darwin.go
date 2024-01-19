package wg_vpn

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

	"github.com/kloudlite/kl/pkg/functions"
	"github.com/miekg/dns"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

const (
	ifName string = "utun2464"
)

func getCurrentDns() ([]string, error) {
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")

	if err != nil {
		return nil, err
	}

	return config.Servers, nil
}

func ipRouteAdd(ip string, interfaceIp string, deviceName string, verbose bool) error {
	return execCmd(fmt.Sprintf("route -n add -net %s %s", ip, interfaceIp), verbose)
}

func getNetworkServices(verbose bool) ([]string, error) {
	output, err := exec.Command("networksetup", "-listallnetworkservices").Output()
	if err != nil {
		functions.Log(string(output))
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
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		dnsServers = append(dnsServers, line)
	}
	return dnsServers, nil
}

func SetDeviceIp(deviceIp net.IPNet, _ string, verbose bool) error {
	return execCmd(fmt.Sprintf("ifconfig %s %s %s", ifName, deviceIp.IP.String(), deviceIp.IP.String()), verbose)
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
	output, err := exec.Command("pgrep", "-f", "kl vpn start-fg").Output()
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
	return nil
}
