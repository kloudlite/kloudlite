package boxpkg

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"

	dockerclient "github.com/docker/docker/client"
	cl "github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
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

	env *cl.Env

	// klConfig *cl.KLFileType

	configFolder string
	userHomeDir  string
}

type BoxClient interface {
	SyncProxy(config ProxyConfig) error
	StopAll() error
	Stop() error
	Start(*cl.KLFileType) error
	Ssh() error
	Reload() error
	PrintBoxes([]Cntr) error
	ListAllBoxes() ([]Cntr, error)
	Info() error
	Exec([]string, io.Writer) error
}

func (c *client) Context() context.Context {
	return c.cmd.Context()
}

func NewClient(cmd *cobra.Command, args []string) (BoxClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())

	if err != nil {
		return nil, functions.NewE(err)
	}

	foreground := fn.ParseBoolFlag(cmd, "foreground")
	verbose := fn.ParseBoolFlag(cmd, "verbose")
	cwd, _ := os.Getwd()

	hash := md5.New()
	hash.Write([]byte(cwd))
	contName := fmt.Sprintf("klbox-%s", fmt.Sprintf("%x", hash.Sum(nil))[:8])

	env, err := cl.EnvOfPath(cwd)
	if err != nil {
		return nil, functions.NewE(err)
	}

	configFolder, err := cl.GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	userHomeDir, err := cl.GetUserHomeDir()
	if err != nil {
		return nil, functions.NewE(err)
	}

	return &client{
		cli:           cli,
		cmd:           cmd,
		args:          args,
		foreground:    foreground,
		verbose:       verbose,
		cwd:           cwd,
		containerName: contName,
		env:           env,
		// klConfig:      klConfig,
		configFolder: configFolder,
		userHomeDir:  userHomeDir,
	}, nil
}
