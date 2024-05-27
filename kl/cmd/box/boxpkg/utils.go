package boxpkg

import (
	"net"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

type Container struct {
	Name string
	Path string
}

func (c *client) getVPNContainer() (Container, error) {
	defCr := Container{}

	crlist, err := c.cli.ContainerList(c.cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: "kl-box-vpn=true"},
		),
		All: true,
	})
	if err != nil {
		return defCr, err
	}

	if len(crlist) >= 1 {
		if len(crlist[0].Names) >= 1 {

			defCr.Name = crlist[0].Names[0]

			if strings.Contains(defCr.Name, "/") {
				s := strings.Split(defCr.Name, "/")
				if len(s) >= 1 {
					defCr.Name = s[1]
				}
			}

			return defCr, nil
		}

		defCr.Name = crlist[0].ID

		return defCr, nil
	}

	return defCr, nil
}

func (c *client) ensurePublicKey() error {
	sshPath := path.Join(xdg.Home, ".ssh")
	if _, err := os.Stat(path.Join(sshPath, "id_rsa")); os.IsNotExist(err) {
		cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", path.Join(sshPath, "id_rsa"), "-N", "")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *client) ensureCacheExist() error {
	vlist, err := c.cli.VolumeList(c.cmd.Context(), volume.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "kl-box-nix-store=true",
		}),
	})
	if err != nil {
		return err
	}

	if len(vlist.Volumes) == 0 {
		if _, err := c.cli.VolumeCreate(c.cmd.Context(), volume.CreateOptions{
			Labels: map[string]string{
				"kl-box-nix-store": "true",
			},
			Name: "nix-store",
		}); err != nil {
			return err
		}
	}

	vlist, err = c.cli.VolumeList(c.cmd.Context(), volume.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "kl-box-nix-home-cache=true",
		}),
	})

	if err != nil {
		return err
	}

	if len(vlist.Volumes) == 0 {
		if _, err := c.cli.VolumeCreate(c.cmd.Context(), volume.CreateOptions{
			Labels: map[string]string{
				"kl-box-nix-home-cache": "true",
			},
			Name: "nix-home-cache",
		}); err != nil {
			return err
		}
	}

	return nil
}

func GetDockerHostIp() (string, error) {

	// if runtime.GOOS != constants.RuntimeLinux {
	// 	return "host.docker.internal"
	// }

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP.To4().String(), nil
}
