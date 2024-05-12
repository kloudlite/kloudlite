package boxpkg

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) EnsureVpnRunning() error {
	defer c.spinner.Update("starting vpn container")()

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
	d, err := server.EnsureDevice()
	if err != nil {
		return err
	}

	configuration, err := base64.StdEncoding.DecodeString(d.WireguardConfig.Value)
	if err != nil {
		return err
	}

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}

	confPath := path.Join(td, "klboxvpn.conf")

	if err := os.WriteFile(confPath, configuration, os.ModePerm); err != nil {
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

func (c *client) StopVpn() error {
	defer c.spinner.Update("stoping vpn container")()

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
