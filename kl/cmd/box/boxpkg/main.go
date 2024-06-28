package boxpkg

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/kloudlite/kl/domain/apiclient"
	"io"
	"os"

	dockerclient "github.com/docker/docker/client"
	 "github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"
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

	env *fileclient.Env

	configFolder string
	userHomeDir  string
}

type BoxClient interface {
	SyncProxy(config ProxyConfig) error
	StopAll() error
	Stop() error
	Start(*fileclient.KLFileType) error
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
	cwd, _ := os.Getwd()

	hash := md5.New()
	hash.Write([]byte(cwd))
	contName := fmt.Sprintf("klbox-%s", fmt.Sprintf("%x", hash.Sum(nil))[:8])
	klFile, err := fileclient.GetKlFile("")
	if err != nil {
		return nil, functions.NewE(err)
	}
	env, err := fileclient.EnvOfPath(cwd)
	if err != nil && errors.Is(err, fileclient.NoEnvSelected) {
		environment, err := apiclient.GetEnvironment(klFile.AccountName, klFile.DefaultEnv)
		if err != nil {
			return nil, functions.NewE(err)
		}
		env = &fileclient.Env{
			Name:        environment.DisplayName,
			TargetNs:    environment.Spec.TargetNamespace,
			SSHPort:     0,
			ClusterName: environment.ClusterName,
		}
		data, err := fileclient.GetExtraData()
		if err != nil {
			return nil, functions.NewE(err)
		}
		if data.SelectedEnvs == nil {
			data.SelectedEnvs = map[string]*fileclient.Env{
				cwd: env,
			}
		} else {
			data.SelectedEnvs[cwd] = env
		}
		if err := fileclient.SaveExtraData(data); err != nil {
			return nil, functions.NewE(err)
		}
	} else if err != nil {
		return nil, functions.NewE(err)
	}

	configFolder, err := fileclient.GetConfigFolder()
	if err != nil {
		return nil, functions.NewE(err)
	}

	userHomeDir, err := fileclient.GetUserHomeDir()
	if err != nil {
		return nil, functions.NewE(err)
	}

	return &client{
		cli:           cli,
		cmd:           cmd,
		args:          args,
		foreground:    foreground,
		verbose:       flags.IsVerbose,
		cwd:           cwd,
		containerName: contName,
		env:           env,
		configFolder:  configFolder,
		userHomeDir:   userHomeDir,
	}, nil
}
