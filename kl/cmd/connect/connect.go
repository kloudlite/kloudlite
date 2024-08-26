package connect

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "connect",
	Short: "start the wireguard connection",
	Long:  "This command will start the wireguard connection",
	Run: func(cmd *cobra.Command, args []string) {
		if err := startWg(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func startWg(cmd *cobra.Command, args []string) error {

	c, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fn.NewE(err)
	}
	existingContainers, err := c.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", boxpkg.CONT_MARK_KEY, "true")),
			filters.Arg("label", fmt.Sprintf("%s=%s", "wg", "true")),
		),
	})
	if err != nil {
		return fn.NewE(err)
	}

	if len(existingContainers) != 0 {
		containerID := existingContainers[0].ID
		timeOut := 0

		if err := c.ContainerStop(context.Background(), containerID, container.StopOptions{
			Timeout: &timeOut,
		}); err != nil {
			return fn.NewE(fmt.Errorf("failed to stop container: %w", err))
		}

		if err := c.ContainerStart(context.Background(), containerID, container.StartOptions{}); err != nil {
			return fn.NewE(fmt.Errorf("failed to start container: %w", err))
		}
		return nil
	}

	boxClient, err := boxpkg.NewClient(cmd, args)
	if err != nil {
		return fn.NewE(err)
	}
	err = boxClient.StartWgContainer()
	if err != nil {
		return fn.NewE(err)
	}
	return nil
}
