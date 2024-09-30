package box

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
	"strings"
)

var stopAllCmd = &cobra.Command{
	Use:   "stop-all",
	Short: "stop all running boxes",
	Run: func(cmd *cobra.Command, args []string) {

		fn.Logf(text.Yellow("[#] this action will stop all the running workspaces. this will end all current running processes in the container. do you want to do you want to proceed? [Y/n] "))
		if !fn.Confirm(strings.ToUpper("Y"), strings.ToUpper("Y")) {
			return
		}

		if err := stopAllContainers(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopAllContainers() error {
	defer spinner.Client.UpdateMessage("stopping running containers")()

	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fn.NewE(err)
	}
	existingContainers, err := c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", boxpkg.CONT_MARK_KEY, "true")),
			//dockerLabelFilter(CONT_WORKSPACE_MARK_KEY, "true"),
		),
	})
	if err != nil {
		return fn.NewE(err)
	}

	if len(existingContainers) == 0 {
		return nil
	}

	for _, d := range existingContainers {
		timeOut := 0
		if err := c.ContainerStop(context.Background(), d.ID, container.StopOptions{
			Signal:  "SIGKILL",
			Timeout: &timeOut,
		}); err != nil {
			return fn.NewE(err)
		}

		if err := c.ContainerRemove(context.Background(), d.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return fn.NewE(err)
		}
	}
	spinner.Client.Stop()
	return nil
}

func init() {
	stopAllCmd.Aliases = append(stopAllCmd.Aliases, "stopall")
}
