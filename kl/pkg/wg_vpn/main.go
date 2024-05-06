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

	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/flags"
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
	command := exec.Command(flags.CliName, "vpn", "start-fg", "-d", devName)
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
	_ string,
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

	// ps := []wgtypes.PeerConfig{}
	//
	// for i := range cfg.Peers {
	// 	// if i <= 1 {
	// 	// 	continue
	// 	// }
	//
	// 	ps = append(ps, cfg.Peers[i])
	//
	// 	fmt.Printf("\n\n%d-> %+v\n\n", i, cfg.Peers[i])
	// }
	//
	// cfg.Peers = ps

	// return fmt.Errorf("wip")

	if len(cfg.Address) == 0 {
		return errors.New("device ip not found")
	} else if e := SetDeviceIp(cfg.Address[0], interfaceName, verbose); e != nil {
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

		dServers, err := getCurrentDns(verbose)
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

		// if runtime.GOOS != constants.RuntimeDarwin {
		// 	emptydns := []net.IP{}
		// 	cfg.DNS = emptydns
		// }

		return nil
	}(); err != nil {
		return err
	}

	if err := func() error {
		if runtime.GOOS != constants.RuntimeLinux {
			return nil
		}

		if isSystemdReslov {
			return setDnsServer(cfg.DNS[0], interfaceName, verbose)
		}

		return setLinuxDnsServers(cfg.DNS, verbose)
	}(); err != nil {
		return err
	}

	if runtime.GOOS == constants.RuntimeDarwin {
		if err := setDnsServers(func() []net.IPNet {
			ipNet := dnsServers

			matched := false
			for _, i2 := range dnsServers {
				if i2.IP.String() == cfg.DNS[0].String() {
					matched = true
					break
				}
			}

			if !matched {
				ipNet = append([]net.IPNet{{
					IP:   cfg.DNS[0],
					Mask: net.CIDRMask(32, 32),
				}}, ipNet...)

				client.SetDns(func() []string {
					var dns []string
					for _, v := range cfg.DNS {
						dns = append(dns, v.String())
					}
					return dns
				}())

			}

			return ipNet
		}(), "Wi-Fi", verbose); err != nil {
			return err
		}
	}

	// cfg.Peers[0].AllowedIPs = append(cfg.Peers[0].AllowedIPs, dnsServers...)

	//for _, p := range cfg.Config.Peers {
	//	fmt.Println("peers ", p.Endpoint)
	//}

	err = wg.ConfigureDevice(interfaceName, cfg.Config)
	if err != nil {
		return err
	}

	for _, pc := range cfg.Peers {
		for _, i2 := range pc.AllowedIPs {
			err = ipRouteAdd(i2.String(), cfg.Address[0].IP.String(), interfaceName, verbose)
			if err != nil {
				return err
			}
		}
	}

	return err
}
