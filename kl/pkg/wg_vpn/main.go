package wg_vpn

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func IsSystemdReslov() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if err := ExecCmd("systemctl status systemd-resolved", false); err != nil {
		return false
	}

	return true
}

func ExecCmd(cmdString string, verbose bool) error {
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

func StartServiceInBg(devName string, configFolder string) error {
	command := exec.Command("kl", "vpn", "start-fg", "-d", devName)
	err := command.Start()
	if err != nil {
		return err
	}

	err = os.WriteFile(configFolder+"/wgpid", []byte(fmt.Sprintf("%d", command.Process.Pid)), 0644)
	if err != nil {
		fn.PrintError(err)
		return err
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err = ExecCmd(fmt.Sprintf("chown %s %s", usr, configFolder+"/wgpid"),
			false); err != nil {
			fn.PrintError(err)
			return err
		}
	}

	return nil
}

func Configure(
	configuration []byte,
	devName string,
	interfaceName string,
	verbose bool,
) error {

	s := spinner.NewSpinner()
	cfg := wgc.Config{}

	s.Start()
	if verbose {
		fn.Log("[#] validating configuration")
	}
	if e := cfg.UnmarshalText(configuration); e != nil {
		return e
	}

	s.Stop()
	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := SetDeviceIp(cfg.Address[0], devName, verbose); e != nil {
		return e
	}

	wg, err := wgctrl.New()
	if err != nil {
		return err
	}

	if verbose {
		fn.Log("[#] setting up connection")
	}

	dnsServers := make([]net.IPNet, 0)
	isSystemdReslov := IsSystemdReslov()

	if err := func() error {
		if isSystemdReslov {
			return nil
		}

		dServers, err := getCurrentDns()
		if err != nil {
			return err
		}

		dnsServers = func() []net.IPNet {
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

		return nil
	}(); err != nil {
		return err
	}

	if isSystemdReslov || runtime.GOOS == "darwin" {
		if err := setDnsServer(cfg.DNS[0], interfaceName, verbose); err != nil {
			return err
		}
	}

	cfg.Peers[0].AllowedIPs = append(cfg.Peers[0].AllowedIPs, dnsServers...)

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		return err
	}

	for _, i2 := range cfg.Peers[0].AllowedIPs {
		err = ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), interfaceName, verbose)
		if err != nil {
			return err
		}
	}

	return err
}
