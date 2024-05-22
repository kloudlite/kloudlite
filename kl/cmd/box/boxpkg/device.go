package boxpkg

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/docker/docker/api/types/container"
	"github.com/kloudlite/kl/constants"
	cl "github.com/kloudlite/kl/domain/client"
	proxy "github.com/kloudlite/kl/domain/dev-proxy"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) ensureVpnConnected() error {
	if err := cl.EnsureAppRunning(); err != nil {
		return err
	}

	p, err := proxy.NewProxy(c.verbose, false)
	if err != nil {
		return err
	}

	if !server.CheckDeviceStatus() {
		functions.Warn("starting vpn")
		if err := p.Start(); err != nil {
			return err
		}
	}

	return nil
}

func (c *client) EnsureVpnCntRunning() error {
	if runtime.GOOS == constants.RuntimeLinux {
		return nil
	}

	defer c.spinner.UpdateMessage("starting vpn container")()

	if c.verbose {
		fn.Logf("starting container in: %s", text.Blue(c.cwd))
	}

	_, err := c.getContainer(map[string]string{
		// CONT_NAME_KEY: c.containerName,
		CONT_VPN_MARK_KEY: "true",
	})

	if err == nil {
		return nil
		// fn.Println(cr.Name)
		// if err := c.StopVpn(); err != nil {
		// 	return err
		// }
		//
		// return c.EnsureVpnRunning()
	}

	if err != nil && err != notFoundErr {
		return err
	}

	c.spinner.Stop()

	dc, err := cl.GetDeviceContext()
	if err != nil {
		return err
	}

	if dc.PrivateKey == nil {
		return fmt.Errorf("something went wrong, wireguard private key not found")
	}

	configuration := fmt.Sprintf(`
[Interface]
ListenPort = 51820
Address = %s/32
PrivateKey = %s 

[Peer]
PublicKey = %s
AllowedIPs = 100.64.0.0/10
	`, dc.DeviceIp.To4().String(), dc.PrivateKey, dc.HostPublicKey)

	// configuration, err := base64.StdEncoding.DecodeString(d.WireguardConfig.Value)
	// if err != nil {
	// 	return err
	// }

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}

	confPath := path.Join(td, "klboxvpn.conf")

	if err := os.WriteFile(confPath, []byte(configuration), os.ModePerm); err != nil {
		return err
	}

	defer func() {
		os.RemoveAll(td)
	}()

	if err := func() error {
		args := []string{
			"--network", "host",
			"--privileged",
			"-v", fmt.Sprintf("%s:/config/wg_confs/klboxvpn.conf", confPath),
			"-p", fmt.Sprintf("%d:51820/udp", constants.ContainerVpnPort),
			VpnImageName,
		}

		if err := c.runContainer(ContainerConfig{
			Name: "kl-vpn-container",
			labels: map[string]string{
				CONT_VPN_MARK_KEY: "true",
			},
			args:      args,
			imageName: VpnImageName,
		}); err != nil {
			return err
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil

}

func (c *client) StopContVpn() error {
	if runtime.GOOS == constants.RuntimeLinux {
		return nil
	}

	defer c.spinner.UpdateMessage("stoping vpn container")()

	cr, err := c.getContainer(map[string]string{
		// CONT_NAME_KEY: c.containerName,
		CONT_VPN_MARK_KEY: "true",
	})

	if err != nil {
		return err
	}

	if err := c.cli.ContainerStop(c.Context(), cr.Name, container.StopOptions{}); err != nil {
		return fmt.Errorf("error stoping container: %s", err)
	}

	if err := c.cli.ContainerRemove(c.Context(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	return nil
}
