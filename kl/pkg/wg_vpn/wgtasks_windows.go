package wg_vpn

import (
	"errors"
	"net"
)

func getCurrentDns(verbose bool) ([]string, error) {
	return []string{}, nil
}

func SetDeviceIp(ip net.IPNet, deviceName string, _ bool) error {
	return errors.New("this command is not available for windows, will be available soon")
}

func StartService(_ string, verbose bool) error {
	return errors.New("this command is not available for windows, will be available soon")
}

func ipRouteAdd(ip string, interfaceIp string, devName string, verbose bool) error {
	return errors.New("this command is not available for windows, will be available soon")
}

func StopService(verbose bool) error {
	return errors.New("this command is not available for windows, will be available soon")
}

func setDnsServers(_ []net.IP, _ string, _ bool) error {
	return errors.New("this command is not available for windows, will be available soon")
}
