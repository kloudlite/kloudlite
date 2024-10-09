package cluster

import (
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/k3s"
	"github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
	"time"
)

var DownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stops the k3s server",
	Long:  `Stops the k3s server`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := StopK3sServer(cmd); err != nil {
			functions.PrintError(err)
			return
		}
	},
}

func StopK3sServer(cmd *cobra.Command) error {
	defer spinner.Client.UpdateMessage("stopping k3s server")()
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	crlist, err := cli.ContainerList(cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", k3s.CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "kl-k3s", "true")),
		),
		All: true,
	})
	if err != nil {
		return err
	}

	if len(crlist) == 0 {
		return nil
	}

	k3sclient, err := k3s.NewClient()
	if err != nil {
		return err
	}

	if err := k3sclient.DeletePods(); err != nil {
		return err
	}

	<-time.After(2 * time.Second)
	for _, c := range crlist {
		if err := cli.ContainerStop(cmd.Context(), c.ID, container.StopOptions{}); err != nil {
			return err
		}
		if err := cli.ContainerRemove(cmd.Context(), c.ID, container.RemoveOptions{}); err != nil {
			return err
		}
	}

	return nil
}
