package boxpkg

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/kloudlite/kl/cmd/box/boxpkg/hashctrl"
	"github.com/kloudlite/kl/domain/apiclient"
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
			err = hashctrl.SyncBoxHash(c.cwd)
			if err != nil {
				return fn.NewE(err)
			}
		}
	} else {
		klconfHash, err := hashctrl.GenerateKLConfigHash(c.klfile)
		if err != nil {
			return fn.NewE(err)
		}
		if klconfHash != boxHash.KLConfHash {
			err = hashctrl.SyncBoxHash(c.cwd)
			if err != nil {
				return functions.NewE(err)
			}
		}
	}

	_, err = c.startContainer()
	if err != nil {
		return fn.NewE(err)
	}

	if err = c.SyncProxy(ProxyConfig{
		ExposedPorts:        c.klfile.Ports,
		TargetContainerPath: c.cwd,
	}); err != nil {
		return fn.NewE(err)
	}

	vpnCfg, err := apiclient.GetAccVPNConfig(c.klfile.AccountName)
	if err != nil {
		return functions.NewE(err)
	}

	err = c.SyncVpn(vpnCfg.WGconf)
	if err != nil {
		return functions.NewE(err)
	}

	spinner.Client.Pause()
	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(c.env.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	return nil
}
