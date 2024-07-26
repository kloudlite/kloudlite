package boxpkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"os"
	"strconv"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

var errContainerNotStarted = fmt.Errorf("container not started")

func (c *client) Start() error {
	defer spinner.Client.UpdateMessage("initiating container please wait")()

	if err := c.ensureKloudliteNetwork(); err != nil {
		return fn.NewE(err)
	}

	if err := c.ensureImage(GetImageName()); err != nil {
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
	_, err = c.startContainer(boxHash.KLConfHash)
	if err != nil {
		return fn.NewE(err)
	}

	if err = c.SyncProxy(ProxyConfig{
		ExposedPorts:        c.klfile.Ports,
		TargetContainerPath: c.cwd,
	}); err != nil {
		return fn.NewE(err)
	}

	vpnCfg, err := c.apic.GetAccVPNConfig(c.klfile.AccountName)
	if err != nil {
		return functions.NewE(err)
	}

	if err = c.SyncVpn(vpnCfg.WGconf); err != nil {
		return functions.NewE(err)
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

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(c.env.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	return nil
}
