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

func getSelectedDevice() (string, error) {

	devName, err := client.GetDeviceContext()
	if err != nil {
		return "", err
	}
	return devName.DeviceName, nil

}

func startConfiguration(verbose bool) error {

	deviceName, err := getSelectedDevice()
	if err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		return configureDarwin(deviceName, verbose)
	}
	return configure(deviceName, deviceName, verbose)
}

func configure(
	devName string,
	interfaceName string,
	verbose bool,
) error {

	s := spinner.NewSpinner()
	cfg := wgc.Config{}

	device, err := server.GetDevice(fn.MakeOption("deviceName", devName))
	if err != nil {
		return err
	}
	// time.Sleep(time.Second * 2)
	if device.Spec.ActiveNamespace == "" {
		return errors.New(fmt.Sprintf("no env name found for device %s, please use env using kl env switch\n", devName))
	}
	if len(device.Spec.Ports) == 0 {
		return errors.New(fmt.Sprintf("no ports found for device %s, please export ports using kl vpn expose\n", devName))
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

	dServers, err := getCurrentDns()
	if err != nil {
		return err
	}

	dnsServers := func() []net.IPNet {

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

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		fn.Log("failed to configure device: %v", err)
	}

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
