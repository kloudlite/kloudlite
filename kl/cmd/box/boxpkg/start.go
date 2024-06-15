package boxpkg

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/kloudlite/kl/constants"
	cl "github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/kloudlite/kl/pkg/wg_vpn/wgc"
)

var errContainerNotStarted = fmt.Errorf("container not started")

func (c *client) Start() error {
	defer spinner.Client.UpdateMessage("initiating container please wait")()

	if c.verbose {
		fn.Logf("starting container in: %s", text.Blue(c.cwd))
	}

	cr, err := c.GetContainer(map[string]string{
		CONT_NAME_KEY: c.containerName,
		CONT_MARK_KEY: "true",
	})
	if err != nil && err != NotFoundErr {
		return err
	}

	if err == nil {
		spinner.Client.Stop()
		crPath := cr.Labels[CONT_PATH_KEY]

		fn.Logf("container %s already running in %s", text.Yellow(cr.Name), text.Blue(crPath))

		if err := c.Stop(); err != nil {
			return err
		}

		return c.Start()
	}

	if err := c.ensurePublicKey(); err != nil {
		return err
	}

	if err := c.ensureCacheExist(); err != nil {
		return err
	}

	// local setup
	// kConf, err := server.LoadDevboxConfig()
	// if err != nil {
	// 	return err
	// }

	spinner.Client.Stop()
	if err := cl.EnsureAppRunning(); err != nil {
		return err
	}

	d, err := server.EnsureDevice()
	if err != nil {
		return err
	}

	localEnv, err := cl.EnvOfPath(c.cwd)
	if err != nil {
		return err
	}

	// e, err := server.GetEnv(localEnv.Name)
	// if err != nil {
	// 	return err
	// }

	configuration, err := base64.StdEncoding.DecodeString(d.WireguardConfig.Value)
	if err != nil {
		return err
	}

	cfg := wgc.Config{}
	err = cfg.UnmarshalText(configuration)
	if err != nil {
		return err
	}

	spinner.Client.Start()

	c.ensureVpnConnected()

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}
	if err := userOwn(td); err != nil {
		return err
	}

	defer func() {
		os.RemoveAll(td)
	}()

	if err := func() error {
		// conf, err := kConf.ToJson()
		// if err != nil {
		// 	return err
		// }

		homeDir, err := cl.GetUserHomeDir()
		if err != nil {
			return err
		}

		sshPath := path.Join(homeDir, ".ssh", "id_rsa.pub")

		akByte, err := os.ReadFile(sshPath)
		if err != nil {
			return err
		}

		ak := string(akByte)

		akTmpPath := path.Join(td, "authorized_keys")

		akByte, err = os.ReadFile(path.Join(homeDir, ".ssh", "authorized_keys"))
		if err == nil {
			ak += fmt.Sprint("\n", string(akByte))
		}

		// for wsl
		if err := func() error {
			if runtime.GOOS != constants.RuntimeLinux {
				return nil
			}

			usersPath := "/mnt/c/Users"
			_, err := os.Stat(usersPath)
			if err != nil {
				return nil
			}

			de, err := os.ReadDir(usersPath)
			if err != nil {
				fn.PrintError(err)
				return nil
			}

			for _, de2 := range de {
				pth := path.Join(usersPath, de2.Name(), ".ssh", "id_rsa.pub")
				if _, err := os.Stat(pth); err != nil {
					continue
				}

				b, err := os.ReadFile(pth)
				if err != nil {
					return err
				}

				ak += fmt.Sprint("\n", string(b))
			}

			return nil
		}(); err != nil {
			return err
		}

		if err := writeOnUserScope(akTmpPath, []byte(ak)); err != nil {
			return err
		}

		args := []string{}

		switch runtime.GOOS {
		case constants.RuntimeWindows:
			// fn.Warn("docker support inside container not implemented yet")
		default:
			args = append(args, "-v", "/var/run/docker.sock:/var/run/docker.sock:ro")
		}

		configFolder, err := cl.GetConfigFolder()
		if err != nil {
			return err
		}

		// hostIp, err := GetDockerHostIp()
		// if err != nil {
		// 	return err
		// }

		if len(cfg.DNS) > 0 {
			args = append(args, []string{
				// "--dns", "127.0.0.2",
				// "--dns", "1.1.1.1",
				// "--dns-search", fmt.Sprintf("%s.svc.%s.local", e.Spec.TargetNamespace, e.ClusterName),
			}...)
		}

		sshPort, err := cl.GetAvailablePort()
		if err != nil {
			return err
		}

		localEnv.SSHPort = sshPort

		if err := cl.SelectEnv(*localEnv); err != nil {
			return err
		}

		// s, err := proxy.GetHostIp()
		// if err != nil {
		// 	return err
		// }

		mountBindFlag := "rw"
		if runtime.GOOS == constants.RuntimeLinux {
			mountBindFlag = "z"
		}

		args = append(args, []string{
			"-v", fmt.Sprintf("%s:/tmp/ssh2/authorized_keys:%s", akTmpPath, mountBindFlag),
			"-v", fmt.Sprintf("kl-home-cache:/home:%s", mountBindFlag),
			"-v", fmt.Sprintf("kl-nix-store:/nix:%s", mountBindFlag),
			"-v", fmt.Sprintf("%s:/home/kl/workspace:z", c.cwd),
			"-v", fmt.Sprintf("%s:/home/kl/.cache/.kl:z", configFolder),
			"-e", fmt.Sprintf("SSH_PORT=%d", sshPort),
			"-e", fmt.Sprintf("KL_WORKSPACE=%s", c.cwd),
			"--add-host=box:127.0.0.1",
			// fmt.Sprintf("--add-host=%s.device.local:%s", d.Metadata.Name, s),
			"-p", fmt.Sprintf("%d:%d", sshPort, sshPort),
			GetImageName(),
		}...)

		if err := c.runContainer(ContainerConfig{
			imageName: GetImageName(),
			Name:      c.containerName,
			trackLogs: true,
			labels: map[string]string{
				CONT_NAME_KEY: c.containerName,
				CONT_PATH_KEY: c.cwd,
				CONT_MARK_KEY: "true",
			},
			args: args,
		}); err != nil {
			c.Stop()
			return err
		}

		if c.cmd.Name() == "start" {
			fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))
		}

		return nil
	}(); err != nil {
		return err
	}

	return nil
}
