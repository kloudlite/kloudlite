package vpn

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/kloudlite/kl/domain/server"

	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"

	"github.com/kloudlite/kl/domain/client"
)

func connect(verbose bool, options ...fn.Option) error {

	client.SetLoading(true)

	defer func() {
		client.SetLoading(false)
	}()

	if err := func() error {

		f, err := client.GetConfigFolder()
		if err != nil {
			return err
		}

		device, err := server.EnsureDevice(options...)
		if err != nil {
			return err
		}

		if device.WireguardConfig.Value == "" {
			return errors.New("no wireguard config found, please try again in few seconds")
		}

		configuration, err := base64.StdEncoding.DecodeString(device.WireguardConfig.Value)
		if err != nil {
			return err
		}

		// client.GetDeviceDns()
		dc, err := client.GetDeviceContext()
		if err != nil {
			return err
		}

		cfg := wgc.Config{}

		if verbose {
			fn.Log("[#] validating configuration")
		}
		if e := cfg.UnmarshalText(configuration); e != nil {
			return e
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

		pth := path.Join(f, fmt.Sprintf("%s.conf", ifName))
		if err := os.WriteFile(pth, configuration, os.ModePerm); err != nil {
			return err
		}

		if _, err := exec.LookPath("wireguard"); err != nil {
			return fmt.Errorf("can't find wireguard in path, please ensure it's installed. installation link %s", text.Blue("https://www.wireguard.com/install"))
		}

		if _, err := fn.WinSudoExec(fmt.Sprintf("%s /installtunnelservice %s", "wireguard", pth), nil); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	data, err := client.GetExtraData()
	if err != nil {
		return err
	}
	data.VpnConnected = true
	if err := client.SaveExtraData(data); err != nil {
		return err
	}
	return nil
}

func disconnect(verbose bool) error {

	if _, err := fn.WinSudoExec(fmt.Sprintf("%s /uninstalltunnelservice %s", "wireguard", ifName), map[string]string{"PATH": os.Getenv("PATH")}); err != nil {
		return err
	}

	data, err := client.GetExtraData()
	if err != nil {
		return err
	}
	data.VpnConnected = false
	if err := client.SaveExtraData(data); err != nil {
		return err
	}

	return nil
}
