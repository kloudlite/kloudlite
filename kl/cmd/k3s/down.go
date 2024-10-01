package k3s

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/k3s"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stops the k3s server",
	Long:  `Stops the k3s server`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := stopK3sServer(cmd); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func stopK3sServer(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("stopping k3s server")()
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	crlist, err := cli.ContainerList(cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", k3s.CONT_MARK_KEY, "true")),
		),
		All: true,
	})
	if err != nil {
		return err
	}

	timeOut := 0
	for _, c := range crlist {
		if err := cli.ContainerStop(cmd.Context(), c.ID, container.StopOptions{
			Signal:  "SIGKILL",
			Timeout: &timeOut,
		}); err != nil {
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
