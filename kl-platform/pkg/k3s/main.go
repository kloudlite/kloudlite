package k3s

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

type k3s struct {
	ctx context.Context
}

type K3s interface {
	StartServer(args []string) error
}

func NewK3s(ctx context.Context) K3s {
	return &k3s{
		ctx: ctx,
	}
}

func (k *k3s) StartServer(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	k3sArgs := []string{"server"}
	for _, v := range args {
		k3sArgs = append(k3sArgs, fmt.Sprintf("--%s", v))
	}

	ctx, cf := signal.NotifyContext(k.ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	cmd := exec.CommandContext(ctx, path.Join(dir, "k3s"), k3sArgs...)
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fn.Log(text.Blue("Kubernetes cluster running with the Kloudlite platform already installed"))

	if flags.IsVerbose {
		fn.Log("Running command:", text.Blue(cmd.String()))

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return cmd.Run()
	}

	b, err := cmd.Output()
	if err != nil {
		fmt.Println(string(b))
		return err
	}

	return nil
}
