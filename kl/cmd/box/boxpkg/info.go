package boxpkg

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/table"
	"github.com/kloudlite/kl/pkg/ui/text"
	"strings"
)

func (c *client) Info() error {

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

	sshPort := cr.Labels[SSH_PORT_KEY]
	fn.Println()

	table.KVOutput("User:", "kl", true)

	table.KVOutput("Name:", strings.Join(cr.Names, ", "), true)
	table.KVOutput("State:", cr.State, true)
	table.KVOutput("Path:", c.cwd, true)
	table.KVOutput("SSH Port:", sshPort, true)

	fn.Logf("%s %s %s\n", text.Bold("command:"), text.Blue("ssh"), text.Blue(strings.Join([]string{fmt.Sprintf("kl@%s", getDomainFromPath(c.cwd)), "-p", fmt.Sprint(sshPort), "-oStrictHostKeyChecking=no"}, " ")))

	fn.Logf("%s %s\n", text.Bold("vscode:"), text.Blue(fmt.Sprintf("vscode://vscode-remote/ssh-remote+kl@%s:%s/home/kl/workspace", getDomainFromPath(c.cwd), sshPort)))
	return nil
}
