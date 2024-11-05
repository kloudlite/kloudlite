package boxpkg

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	fn "github.com/kloudlite/kl/pkg/functions"
)

func (c *client) Exec(command []string, out io.Writer) error {
	if len(command) == 0 {
		return fn.Error("command not provided")
	}

	existingContainers, err := c.cli.ContainerList(context.Background(), container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "kloudlite=true"),
			filters.Arg("label", "workspacebox=true"),
			filters.Arg("label", fmt.Sprintf("working_dir=%s", c.cwd)),
		),
	})
	if err != nil {
		return fn.Error("failed to list containers")
	}
	if len(existingContainers) == 0 {
		return fn.Error("container not running")
	}

	execIDResp, err := c.cli.ContainerExecCreate(context.Background(), existingContainers[0].ID, container.ExecOptions{
		Cmd:          command,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
	})
	if err != nil {
		return fn.NewE(err, "failed to create exec")
	}

	execID := execIDResp.ID
	if execID == "" {
		return fn.Error("exec ID empty")
	}

	resp, err := c.cli.ContainerExecAttach(context.Background(), execID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return fn.NewE(err)
	}

	defer resp.Close()
	if out == nil {
		out = os.Stdout
	}

	_, err = io.Copy(out, resp.Reader)
	if err != nil && err != io.EOF {
		return fn.NewE(err)
	}

	_, err = c.getExecExitCode(context.Background(), execID)
	return fn.NewE(err)
}

func (c *client) getExecExitCode(ctx context.Context, execID string) (int, error) {
	for {
		inspectResp, err := c.cli.ContainerExecInspect(ctx, execID)
		if err != nil {
			return 0, err
		}

		if !inspectResp.Running {
			return inspectResp.ExitCode, nil
		}
	}
}
