package boxpkg

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/kloudlite/kl/pkg/dockercli"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func (c *client) Stop() error {
	s := spinner.NewSpinner("stopping container please wait")
	if !c.verbose {
		s.Start()
		defer s.Stop()
	}

	cr, err := c.getContainer()
	if err != nil {
		return err
	}

	if cr.Name == "" {
		return fmt.Errorf("no running container found")
	}

	if c.verbose {
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

	if c.verbose {
		fn.Logf("stopped container of: %s", text.Blue(cr.Path))
	}
	return nil
}
