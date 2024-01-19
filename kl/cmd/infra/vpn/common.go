package vpn

import (
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/lib/wgc"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func getDeviceSelect() (*server.Device, error) {

	devName, err := client.CurrentInfraDeviceName()
	if err != nil {
		return nil, err
	}

	devices, err := server.ListInfraDevices()
	if err != nil {
		return nil, err
	}

	for _, d := range devices {
		if d.Metadata.Name == devName {
			return &d, err
		}
	}
	return nil, errors.New("please select an infra context first using \"kl infra vpn switch\"")

}

func startConfiguration(verbose bool) error {
	devices, err := server.ListInfraDevices()
	if err != nil {
		return err
	}
	if len(devices) == 0 {
		return errors.New("no infra Devices found")
	}
	device, err := getDeviceSelect()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		return configureDarwin(device.Metadata.Name, verbose)
	}

	return configure(device.Metadata.Name, device.Metadata.Name, verbose)
}

func configure(
	devName string,
	interfaceName string,
	verbose bool,
) error {

	s := spinner.NewSpinner()
	cfg := wgc.Config{}

	device, err := server.GetInfraDevice(fn.MakeOption("deviceName", devName))
	if err != nil {
		return err
	}

	if device.Spec.ActiveNamespace == "" {
		return errors.New(fmt.Sprintf("no env name found for device %s, please use env using kl env switch\n", devName))
	}
	if len(device.Spec.Ports) == 0 {
		return errors.New(fmt.Sprintf("no ports found for device %s, please export ports using kl infra vpn expose\n", devName))
	}
	if device.WireguardConfig == nil {
		return errors.New("no wireguard config found")
	}

	configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
	if err != nil {
		return err
	}

	s.Start()
	if verbose {
		fn.Log("[#] validating configuration")
	}
	if e := cfg.UnmarshalText([]byte(configuration)); e != nil {
		return e
	}
	s.Stop()

	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := setDeviceIp(cfg.Address[0], devName, verbose); e != nil {
		return e
	}

	wg, err := wgctrl.New()
	if err != nil {
		return err
	}

	if verbose {
		fn.Log("[#] setting up connection")
	}

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		fn.Log("failed to configure device: %v", err)
	}

	dServers, err := getCurrentDns()
	if err != nil {
		return err
	}

	dnsServers := func() []net.IPNet {
		//	var ipNet []net.IPNet
		//	for _, v := range dServers {
		//		ipNet = append(ipNet, net.IPNet{
		//			IP:   net.ParseIP(v),
		//			Mask: net.CIDRMask(32, 32),
		//		})
		//	}
		var ipNet []net.IPNet
		for _, v := range dServers {
			ip := net.ParseIP(v)
			if ip == nil {
				continue
			}
			in := net.IPNet{
				IP: ip,
				Mask: func() net.IPMask {
					if ip.To4() != nil {
						return net.CIDRMask(32, 32)
					}
					return net.CIDRMask(128, 128)
				}(),
			}
			ipNet = append(ipNet, in)
		}

		return ipNet
	}()

	emptydns := []net.IP{}
	cfg.DNS = emptydns

	cfg.Peers[0].AllowedIPs = append(cfg.Peers[0].AllowedIPs, dnsServers...)

	for _, i2 := range cfg.Peers[0].AllowedIPs {
		err = ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), interfaceName, verbose)
		if err != nil {
			return err
		}
	}

	return err
}

func execCmd(cmdString string, verbose bool) error {
	r := csv.NewReader(strings.NewReader(cmdString))
	r.Comma = ' '
	cmdArr, err := r.Read()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdArr[0], cmdArr[1:]...)
	if verbose {
		fn.Log("[#] " + strings.Join(cmdArr, " "))
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	// s.Start()
	err = cmd.Run()
	// s.Stop()
	return err
}
