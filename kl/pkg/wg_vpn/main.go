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
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	err = cmd.Run()
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
	interfaceName string,
	verbose bool,
) error {

	stopSpinner := spinner.Client.UpdateMessage("validating configuration")
	cfg := wgc.Config{}

	if verbose {
		fn.Log("[#] validating configuration")
	}
	if e := cfg.UnmarshalText(configuration); e != nil {
		return e
	}

	stopSpinner()

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

	if err := func() error {
		if runtime.GOOS != constants.RuntimeLinux {
			return nil
		}

		if len(cfg.Address) > 0 {
			dc, err := client.GetDeviceContext()
			if err != nil {
				return err
			}

			priv := dc.PrivateKey

			if priv == nil {
				_, priv, err = GenerateWgKeys()
				if err != nil {
					return err
				}

				dc.PrivateKey = priv
			}

			pub, err := GeneratePublicKey(string(priv))
			if err != nil {
				return err
			}

			if len(pub) < 32 {
				fmt.Println("wrong public key length")
			}
			var pubBuff [32]byte
			copy(pubBuff[:], pub[:32])

			hostPublicKey, err := GeneratePublicKey(cfg.PrivateKey.String())
			if err != nil {
				return err
			}
			dc.HostPublicKey = hostPublicKey

			cfg.Peers = append(cfg.Peers, wgtypes.PeerConfig{
				PublicKey: pubBuff,
				Endpoint: &net.UDPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: constants.ContainerVpnPort,
				},
				AllowedIPs: []net.IPNet{{
					IP:   cfg.Address[0].IP,
					Mask: net.CIDRMask(32, 32),
				}},
			})

			dc.DeviceIp = cfg.Address[0].IP
			cfg.Address = []net.IPNet{}

			if err := client.WriteDeviceContext(dc); err != nil {
				return err
			}
		}

		return nil
	}(); err != nil {
		return err
	}

	// fn.Log(cfg.PrivateKey.String(), cfg.Address[0].IP.To4().String())

	if err := SetDnsServers(cfg.DNS, interfaceName, verbose); err != nil {
		return err
	}

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

	if err != nil {
		return err
	}

	if len(cfg.DNS) > 0 {
		return client.SetDeviceDns(cfg.DNS[0].To4().String())
	}

	return nil
}
