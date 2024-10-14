package boxpkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/cmd/cluster"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/domain/fileclient"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

var errContainerNotStarted = fn.Errorf("container not started")

func (c *client) Start() error {
	defer spinner.Client.UpdateMessage("initiating container please wait")()

	if err := c.k3s.EnsureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	if err := c.ensureImage(constants.GetBoxImageName()); err != nil {
		return fn.NewE(err)
	}

	boxHash, err := hashctrl.BoxHashFile(c.cwd)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = hashctrl.SyncBoxHash(c.apic, c.fc, c.cwd)
			if err != nil {
				return fn.NewE(err)
			}
			return c.Start()
		}
	} else {
		klconfHash, err := hashctrl.GenerateKLConfigHash(c.klfile)
		if err != nil {
			return fn.NewE(err)
		}
		if klconfHash != boxHash.KLConfHash {
			err = hashctrl.SyncBoxHash(c.apic, c.fc, c.cwd)
			if err != nil {
				return functions.NewE(err)
			}
		}
	}
	if boxHash == nil {
		boxHash, err = hashctrl.BoxHashFile(c.cwd)
		if err != nil {
			return fn.NewE(err)
		}
	}

	// if err = c.k3s.CreateClustersTeams(c.klfile.TeamName); err != nil {
	// 	return fn.NewE(err)
	// }

	data, err := fileclient.GetExtraData()
	if err != nil {
		return fn.NewE(err)
	}
	if data.SelectedTeam != c.klfile.TeamName {
		functions.Logf(text.Yellow(fmt.Sprintf("[#] this will switch your main team context from %s to %s. do you want to proceed? [Y/n] ", data.SelectedTeam, c.klfile.TeamName)))
		if !functions.Confirm("y", "y") {
			return nil
		}

		data.SelectedTeam = c.klfile.TeamName

		if err := fileclient.SaveExtraData(data); err != nil {
			return functions.NewE(err)
		}

		if err = cluster.StopK3sServer(c.cmd); err != nil {
			return functions.NewE(err)
		}

		_, err = c.apic.GetClusterConfig(c.klfile.TeamName)
		if err != nil {
			return err
		}

		_, err = c.apic.GetAccVPNConfig(c.klfile.TeamName)
		if err != nil {
			return err
		}

	}

	_, err = c.startContainer(boxHash.KLConfHash)
	if err != nil {
		return fn.NewE(err)
	}

	if c.env.SSHPort == 0 {
		existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
			Filters: filters.NewArgs(
				dockerLabelFilter(CONT_MARK_KEY, "true"),
				dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
				dockerLabelFilter(CONT_PATH_KEY, c.cwd),
			),
			All: true,
		})
		if err != nil {
			return fn.NewE(err)
		}

		if len(existingContainers) == 0 {
			return fn.Error("no container running in current directory")
		}

		cr := existingContainers[0]

		c.env.SSHPort, err = strconv.Atoi(cr.Labels[SSH_PORT_KEY])
		if err != nil {
			return fn.NewE(err)
		}
	}

	if data.SelectedEnvs[c.cwd].SSHPort == 0 {
		data.SelectedEnvs[c.cwd].SSHPort = c.env.SSHPort
		if err := fileclient.SaveExtraData(data); err != nil {
			return fn.NewE(err)
		}
	}

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(c.env.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))
	fn.Logf("%s %s\n", text.Bold("vscode:"), text.Blue(fmt.Sprintf("vscode://vscode-remote/ssh-remote+kl@%s:%s/workspace", getDomainFromPath(c.cwd), fmt.Sprint(c.env.SSHPort))))

	return nil
}

func (c *client) StartClusterContainer() error {
	defer spinner.Client.UpdateMessage("starting k3s cluster")()
	_, err := c.apic.GetClusterConfig(c.klfile.TeamName)
	if err != nil {
		return fn.NewE(err)
	}
	err = c.EnsureK3SCluster(c.klfile.TeamName)
	if err != nil {
		return fn.NewE(err)
	}
	config, err := c.fc.GetClusterConfig(c.klfile.TeamName)
	if config.Installed {
		return nil
	}
	err = c.ConnectClusterToTeam(config)
	if err != nil {
		return fn.NewE(err)
	}
	config.Installed = true
	err = c.fc.SetClusterConfig(c.klfile.TeamName, config)
	if err != nil {
		return fn.NewE(err)
	}
	return nil
}

func (c *client) StartWgContainer() error {
	defer spinner.Client.UpdateMessage("starting wireguard")()
	vpnCfg, err := c.apic.GetAccVPNConfig(c.klfile.TeamName)
	if err != nil {
		return fn.NewE(err)
	}

	if err = c.SyncVpn(vpnCfg.WGconf); err != nil {
		return fn.NewE(err)
	}
	return nil
}
