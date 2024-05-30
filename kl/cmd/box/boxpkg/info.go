package boxpkg

import (
	"fmt"
	"strings"

	cl "github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Info(contName string) error {
	cr, err := c.getContainer(map[string]string{
		CONT_MARK_KEY: "true",
		CONT_NAME_KEY: contName,
	})

	if err != nil && err != notFoundErr {
		return err
	}

	if err == notFoundErr {
		fn.Logf("no running container found")
		return nil
	}

	pth := cr.Labels[CONT_PATH_KEY]
	localEnv, err := cl.EnvOfPath(pth)
	if err != nil {
		return err
	}

	table.KVOutput("User:", "kl", true)

	table.KVOutput("Name:", cr.Name, true)
	table.KVOutput("Path:", pth, true)
	table.KVOutput("SSH Port:", localEnv.SSHPort, true)

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(pth)), "-p", fmt.Sprint(localEnv.SSHPort), "-oStrictHostKeyChecking=no"}, " ")))

	return nil
}
