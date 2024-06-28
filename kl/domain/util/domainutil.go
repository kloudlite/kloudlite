package domainutil

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/cmd/box/boxpkg"
	"github.com/kloudlite/kl/domain/envclient"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)



func ConfirmBoxRestart(cwd string) error {
	if envclient.InsideBox() {
		return nil
	}
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	existingContainers, err :=cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=true", boxpkg.CONT_MARK_KEY)),
			filters.Arg("label", fmt.Sprintf("%s=true", boxpkg.CONT_WORKSPACE_MARK_KEY)),
			filters.Arg("label", fmt.Sprintf("%s=%s", boxpkg.CONT_PATH_KEY, cwd)),
		),
		All: true,
	})
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if len(existingContainers) == 0 {
		return nil
	}

	fn.Log(text.Yellow("environments may have been updated. to reflect the changes, do you want to restart the container? [Y/n] "))
	if !fn.Confirm("Y", "Y"){
		return nil
	}
	_, err = fn.Exec("kl box restart", nil)
	if err != nil {
		return err
	}
	return nil
}
