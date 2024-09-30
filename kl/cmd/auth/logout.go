package auth

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/fileclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout from kloudlite",
	Example: `# Logout from kloudlite
{cmd} auth logout`,
	Run: func(cmd *cobra.Command, args []string) {
		fc, err := fileclient.New()
		if err != nil {
			fn.PrintError(err)
			return
		}
		err = stopAllContainers(cmd, args)
		if err != nil {
			fn.PrintError(err)
			return
		}

		if err := fc.Logout(); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopAllContainers(cmd *cobra.Command, args []string) error {
	defer spinner.Client.UpdateMessage("stopping container please wait")()
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	crlist, err := cli.ContainerList(cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("%s=%s", boxpkg.CONT_MARK_KEY, "true")},
		),
		All: true,
	})
	if err != nil {
		return err
	}
	for _, c := range crlist {
		if err := cli.ContainerStop(cmd.Context(), c.ID, container.StopOptions{}); err != nil {
			return err
		}
		if err := cli.ContainerRemove(cmd.Context(), c.ID, container.RemoveOptions{
			Force: true,
		}); err != nil {
			return err
		}
	}
	return nil
}
