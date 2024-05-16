package boxpkg

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"

	dockerclient "github.com/docker/docker/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/spf13/cobra"
)

type client struct {
	cli        *dockerclient.Client
	cmd        *cobra.Command
	args       []string
	foreground bool
	verbose    bool
	cwd        string

	containerName string

	spinner *spinner.Spinner
}

func (c *client) Context() context.Context {
	return c.cmd.Context()
}

func NewClient(cmd *cobra.Command, args []string) (*client, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	foreground := fn.ParseBoolFlag(cmd, "foreground")
	verbose := fn.ParseBoolFlag(cmd, "verbose")
	cwd, _ := os.Getwd()

	hash := md5.New()
	hash.Write([]byte(cwd))
	contName := fmt.Sprintf("kl-box-%x", hash.Sum(nil))

	sp := spinner.NewSpinner2("loading please wait", (foreground || verbose))

	return &client{
		cli:           cli,
		cmd:           cmd,
		args:          args,
		foreground:    foreground,
		verbose:       verbose,
		cwd:           cwd,
		containerName: contName,
		spinner:       sp,
	}, nil
}
