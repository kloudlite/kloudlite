package box

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/adrg/xdg"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/pkg/dockercli"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "get ssh access to the container",
	Run: func(cmd *cobra.Command, args []string) {
		if err := sshBox(cmd, args); err != nil {
			fn.PrintError(err)
			return
		}
	},
}

type Container struct {
	Name string
	Path string
}

func getRunningContainer() (Container, error) {
	ctx := context.Background()

	defCr := Container{}

	cli, err := dockercli.GetClient()
	if err != nil {
		return defCr, err
	}
	defer cli.Close()

	c, err := cli.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "label",
			Value: "kl-box=true",
		}),
	})
	if err != nil {
		return defCr, err
	}

	if len(c) >= 1 {
		if len(c[0].Names) >= 1 {
			defCr.Name = c[0].Names[0]
			defCr.Path = c[0].Labels["kl-box-cwd"]
			return defCr, nil
		}

		defCr.Name = c[0].ID
		defCr.Path = c[0].Labels["kl-box-cwd"]

		return defCr, nil
	}

	return defCr, nil
}

func sshBox(cmd *cobra.Command, args []string) error {
	debug := fn.ParseBoolFlag(cmd, "debug")

	containerName := "kl-box-" + getCwdHash()

	cont, err := getRunningContainer()
	if err != nil {
		return err
	}

	s := spinner.NewSpinner("waiting for container to be ready")
	if cont.Name == "" {
		if err := startBox(cmd, args); err != nil {
			return err
		}

		s.Start()
		time.Sleep(5 * time.Second)
		s.Stop()
	} else if fmt.Sprintf("/%s", containerName) != cont.Name {

		if debug {
			fn.Logf("already running in: %s", text.Blue(cont.Path))
		}

		if err := stopBox(cmd, args); err != nil {
			return err
		}

		if err := startBox(cmd, args); err != nil {
			return err
		}

		s.Start()
		time.Sleep(5 * time.Second)
		s.Stop()
	}

	command := exec.Command("ssh", "kl@localhost", "-p", "1729", "-i", path.Join(xdg.Home, ".ssh", "id_rsa"))

	if debug {
		fn.Log(command.String())
	}

	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	if err := command.Run(); err != nil {
		return fmt.Errorf(("error opening ssh to kl-box container. Please ensure that container is running, or wait for it to start. %s"), err)
	}
	return nil
}

func init() {
	sshCmd.Flags().BoolP("debug", "d", false, "run in debug mode")
}
