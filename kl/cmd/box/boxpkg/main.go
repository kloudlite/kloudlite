package boxpkg

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/kloudlite/kl/domain/apiclient"

	dockerclient "github.com/docker/docker/client"
	"github.com/kloudlite/kl/domain/fileclient"
	"github.com/kloudlite/kl/flags"

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

	fc     fileclient.FileClient
	klfile *fileclient.KLFileType
}

type BoxClient interface {
	SyncProxy(config ProxyConfig) error
	StopAll() error
	Stop() error
	Restart() error
	Start() error
	Ssh() error
	Reload() error
	PrintBoxes([]Cntr) error
	ListAllBoxes() ([]Cntr, error)
	Info() error
	Exec([]string, io.Writer) error

	ConfirmBoxRestart() error
}

func (c *client) Context() context.Context {
	return c.cmd.Context()
}

func NewClient(cmd *cobra.Command, args []string) (BoxClient, error) {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fn.NewE(err)
	}

	fc, err := fileclient.New()
	if err != nil {
		return nil, fn.NewE(err)
	}

	if err != nil {
		return nil, fn.NewE(err)
	}

	foreground := fn.ParseBoolFlag(cmd, "foreground")
	cwd, _ := os.Getwd()

	hash := md5.New()
	hash.Write([]byte(cwd))
	contName := fmt.Sprintf("klbox-%s", fmt.Sprintf("%x", hash.Sum(nil))[:8])

	klFile, err := fc.GetKlFile("")
	if err != nil {
		return nil, fn.NewE(err)
	}

	env, err := fileclient.EnvOfPath(cwd)
	if err != nil && errors.Is(err, fileclient.NoEnvSelected) {
		environment, err := apiclient.GetEnvironment(klFile.AccountName, klFile.DefaultEnv)
		if err != nil {
			return nil, fn.NewE(err)
		}
		env = &fileclient.Env{
			Name:        environment.DisplayName,
			TargetNs:    environment.Spec.TargetNamespace,
			SSHPort:     0,
			ClusterName: environment.ClusterName,
		}
		data, err := fileclient.GetExtraData()
		if err != nil {
			return nil, fn.NewE(err)
		}
		if data.SelectedEnvs == nil {
			data.SelectedEnvs = map[string]*fileclient.Env{
				cwd: env,
			}
		} else {
			data.SelectedEnvs[cwd] = env
		}
		if err := fileclient.SaveExtraData(data); err != nil {
			return nil, fn.NewE(err)
		}
	} else if err != nil {
		return nil, fn.NewE(err)
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
		fc:            fc,
		klfile:        klFile,
	}, nil
}
