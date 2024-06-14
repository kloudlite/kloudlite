package wg_vpn

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

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

	// client.GetDeviceDns()
	dc, err := client.GetDeviceContext()
	if err != nil {
		return err
	}

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

	dc.DeviceDns = func() []string {
		var dns []string

		for _, d := range cfg.DNS {
			dns = append(dns, d.To4().String())
		}

		return dns
	}()

	if err := client.WriteDeviceContext(dc); err != nil {
		return err
	}

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

	return nil
}
