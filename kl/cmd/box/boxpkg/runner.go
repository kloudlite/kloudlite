package boxpkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/constants"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/nxadm/tail"
)

type ContainerConfig struct {
	imageName string
	Name      string
	labels    map[string]string
	args      []string
	trackLogs bool
}

type ContState string

const (
	ContStateExited  ContState = "exited"
	ContStateCreated ContState = "created"
)

type Cntr struct {
	Name   string
	Labels map[string]string
	State  ContState
}

var NotFoundErr = errors.New("container not found")

func (c *client) listContainer(labels map[string]string) ([]Cntr, error) {
	defer spinner.Client.UpdateMessage("fetching existing container")()

	labelArgs := make([]filters.KeyValuePair, 0)

	for k, v := range labels {
		labelArgs = append(labelArgs, filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("%s=%s", k, v)})
	}

	crlist, err := c.cli.ContainerList(c.cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			labelArgs...,
		),
		All: true,
	})
	if err != nil {
		return nil, err
	}

	if len(crlist) == 0 {
		return nil, NotFoundErr
	}

	defCrs := make([]Cntr, 0)

	for _, c2 := range crlist {

		if len(c2.Names) == 0 {
			defCr := Cntr{
				Name:   crlist[0].ID,
				Labels: crlist[0].Labels,
				State:  ContState(c2.State),
			}
			defCrs = append(defCrs, defCr)
			continue
		}

		defCr := Cntr{
			Name:   c2.Names[0],
			Labels: c2.Labels,
			State:  ContState(c2.State),
		}

		if strings.Contains(defCr.Name, "/") {
			s := strings.Split(defCr.Name, "/")
			if len(s) >= 1 {
				defCr.Name = s[1]
			}
		}

		defCrs = append(defCrs, defCr)
	}

	return defCrs, nil
}

func (c *client) GetContainer(labels map[string]string) (*Cntr, error) {
	defer spinner.Client.UpdateMessage("fetching existing container")()

	labelArgs := make([]filters.KeyValuePair, 0)

	for k, v := range labels {
		labelArgs = append(labelArgs, filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("%s=%s", k, v)})
	}

	crlist, err := c.cli.ContainerList(c.cmd.Context(), container.ListOptions{
		Filters: filters.NewArgs(
			labelArgs...,
		),
		All: true,
	})
	if err != nil {
		return nil, err
	}

	if len(crlist) == 0 {
		return nil, NotFoundErr
	}

	if len(crlist[0].Names) >= 1 {
		defCr := Cntr{
			Name:   crlist[0].Names[0],
			Labels: crlist[0].Labels,
			State:  ContState(crlist[0].State),
		}

		if strings.Contains(defCr.Name, "/") {
			s := strings.Split(defCr.Name, "/")
			if len(s) >= 1 {
				defCr.Name = s[1]
			}
		}

		return &defCr, nil
	}

	defCr := Cntr{
		Name:   crlist[0].ID,
		Labels: crlist[0].Labels,
		State:  ContState(crlist[0].State),
	}

	return &defCr, nil
}

func (c *client) runContainer(config ContainerConfig) error {
	defer spinner.Client.UpdateMessage(fmt.Sprintf("trying to start container %s please wait", config.Name))()

	if c.verbose {
		fn.Logf("starting container %s", text.Blue(config.Name))
	}

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}

	if err := userOwn(td); err != nil {
		return err
	}

	// defer func() {
	// 	os.RemoveAll(td)
	// }()

	stdErrPath := path.Join(td, "stderr.log")
	stdOutPath := path.Join(td, "stdout.log")

	if err := func() error {
		dockerArgs := []string{"run"}
		if !c.foreground {
			dockerArgs = append(dockerArgs, "-d")
		}
		dockerArgs = append(dockerArgs, "--name", config.Name)

		if err := writeOnUserScope(stdOutPath, []byte("")); err != nil {
			return err
		}

		if err := writeOnUserScope(stdErrPath, []byte("")); err != nil {
			return err
		}

		labelArgs := make([]string, 0)
		for k, v := range config.labels {
			labelArgs = append(labelArgs, "--label", fmt.Sprintf("%s=%s", k, v))
		}

		dockerArgs = append(dockerArgs,
			"--hostname", "box",
		)

		mountBindFlag := "rw"
		if runtime.GOOS == constants.RuntimeLinux {
			mountBindFlag = "Z"
		}
		if config.trackLogs {
			dockerArgs = append(dockerArgs,
				"-v", fmt.Sprintf("%s:/tmp/stdout.log:%s", stdOutPath, mountBindFlag),
				"-v", fmt.Sprintf("%s:/tmp/stderr.log:%s", stdErrPath, mountBindFlag),
			)
		}

		dockerArgs = append(dockerArgs, labelArgs...)
		dockerArgs = append(dockerArgs, config.args...)

		// var command *exec.Cmd
		// if _, ok := os.LookupEnv("SUDO_USER"); ok {
		// 	sudoUser := []string{"-u", os.Getenv("SUDO_USER"), "docker"}
		// 	sudoUser = append(sudoUser, dockerArgs...)

		// 	command = exec.Command("sudo", sudoUser...)

		// } else {
		// }

		command := exec.Command("docker", dockerArgs...)

		if c.verbose {
			command.Stdout = os.Stdout
		}
		command.Stderr = os.Stderr

		if c.verbose {
			fn.Logf("docker container started with cmd: %s\n", text.Blue(command.String()))
		}

		// if err := fn.ExecCmd(fmt.Sprintf("docker pull %s", config.imageName), nil, c.verbose); err != nil {
		// 	return err
		// }

		if err := command.Run(); err != nil {
			return fmt.Errorf("error running kl-box container [%s]", err.Error())
		}

		return nil
	}(); err != nil {
		return err
	}

	if !config.trackLogs {
		return nil
	}

	timeoutCtx, cf := context.WithTimeout(c.Context(), 1*time.Minute)

	cancelFn := func() {
		defer cf()
	}

	status := make(chan int, 1)
	go func() {
		ok, err := c.readTillLine(timeoutCtx, stdErrPath, "kloudlite-entrypoint:CRASHED", "stderr", true)
		if err != nil {
			fn.PrintError(err)
			status <- 2
			cf()
			return
		}
		if ok {
			status <- 1
		}
	}()

	go func() {
		ok, err := c.readTillLine(timeoutCtx, stdOutPath, "kloudlite-entrypoint: SETUP_COMPLETE", "stdout", true)
		if err != nil {
			fn.PrintError(err)
			status <- 2
			return
		}

		if ok {
			status <- 0
		}
	}()

	select {
	case exitCode := <-status:
		{
			spinner.Client.Stop()
			cancelFn()
			if exitCode != 0 {
				_ = c.Stop()

				c.verbose = true
				c.readTillLine(timeoutCtx, stdOutPath, "kloudlite-entrypoint: SETUP_COMPLETE", "stdout", false)
				c.readTillLine(timeoutCtx, stdErrPath, "kloudlite-entrypoint:CRASHED", "stderr", false)
				return errors.New("failed to start container")
			}

			if c.verbose {
				fn.Log(text.Blue("container started successfully"))
			}
		}
	}

	return nil
}

func (c *client) readTillLine(ctx context.Context, file string, desiredLine, stream string, follow bool) (bool, error) {
	t, err := tail.TailFile(file, tail.Config{Follow: follow, ReOpen: follow, Poll: runtime.GOOS == constants.RuntimeWindows})
	if err != nil {
		return false, err
	}

	for l := range t.Lines {
		if l.Text == desiredLine {
			return true, nil
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
			spinner.Client.UpdateMessage("installing nix packages")
			continue
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			spinner.Client.UpdateMessage("loading please wait")
			continue
		}

		if c.verbose {
			switch stream {
			case "stderr":
				fn.Logf("%s: %s", text.Yellow("[stderr]"), l.Text)
			default:
				fn.Logf("%s: %s", text.Blue("[stdout]"), l.Text)
			}
		}
	}

	return false, nil
}

func writeOnUserScope(fpath string, data []byte) error {
	if err := os.WriteFile(fpath, data, 0o644); err != nil {
		return err
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return err
		}
	}

	return nil
}

func userOwn(fpath string) error {
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return err
		}
	}

	return nil
}
