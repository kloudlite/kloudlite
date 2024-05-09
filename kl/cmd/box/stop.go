package box

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/kloudlite/kl/pkg/dockercli"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "stop running container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

func stopBox(cmd *cobra.Command, _ []string) error {

	debug := fn.ParseBoolFlag(cmd, "debug")

	s := spinner.NewSpinner("stopping container please wait")

	s.Start()
	defer s.Stop()

	cr, err := getRunningContainer()
	if err != nil {
		return err
	}

	if cr.Name == "" {
		return fmt.Errorf("no running container found")
	}

	if debug {
		fn.Logf("stopping container of: %s", text.Blue(cr.Path))
	}

	cli, err := dockercli.GetClient()
	if err != nil {
		return err
	}

	if err := cli.ContainerStop(context.TODO(), cr.Name, container.StopOptions{}); err != nil {
		return fmt.Errorf("error stoping container: %s", err)
	}

	if err := cli.ContainerRemove(context.TODO(), cr.Name, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove container: %s", err)
	}

	if debug {
		fn.Logf("stopped container of: %s", text.Blue(cr.Path))
	}
	return nil
}

func init() {
	stopCmd.Flags().BoolP("debug", "d", false, "run in debug mode")
}
