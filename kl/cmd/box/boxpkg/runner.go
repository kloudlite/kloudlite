package boxpkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	fn "github.com/kloudlite/kl/pkg/functions"
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

var notFoundErr = errors.New("container not found")

func (c *client) listContainer(labels map[string]string) ([]Cntr, error) {
	defer c.spinner.UpdateMessage("fetching existing container")()

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
		return nil, notFoundErr
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

func (c *client) getContainer(labels map[string]string) (*Cntr, error) {
	defer c.spinner.UpdateMessage("fetching existing container")()

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
		return nil, notFoundErr
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
	defer c.spinner.UpdateMessage(fmt.Sprintf("trying to start container %s please wait", config.Name))()

	if c.verbose {
		fn.Logf("starting container %s", text.Blue(config.Name))
	}

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}

	// defer func() {
	// 	os.RemoveAll(td)
	// }()

	stdErrPath := fmt.Sprintf("%s/stderr.log", td)
	stdOutPath := fmt.Sprintf("%s/stdout.log", td)

	if err := func() error {

		dockerArgs := []string{"run"}
		if !c.foreground {
			dockerArgs = append(dockerArgs, "-d")
			dockerArgs = append(dockerArgs, "--name", config.Name)
		}

		stdErrPath := fmt.Sprintf("%s/stderr.log", td)
		stdOutPath := fmt.Sprintf("%s/stdout.log", td)

		if err := os.WriteFile(stdOutPath, []byte(""), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(stdErrPath, []byte(""), os.ModePerm); err != nil {
			return err
		}

		labelArgs := make([]string, 0)
		for k, v := range config.labels {
			labelArgs = append(labelArgs, "--label", fmt.Sprintf("%s=%s", k, v))
		}

		dockerArgs = append(dockerArgs,
			"--hostname", "box",
		)

		if config.trackLogs {
			dockerArgs = append(dockerArgs,
				"-v", fmt.Sprintf("%s:/tmp/stdout.log:rw", stdOutPath),
				"-v", fmt.Sprintf("%s:/tmp/stderr.log:rw", stdErrPath),
			)
		}

		dockerArgs = append(dockerArgs, labelArgs...)
		dockerArgs = append(dockerArgs, config.args...)

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
			c.spinner.Stop()
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

	t, err := tail.TailFile(file, tail.Config{Follow: follow, ReOpen: follow, Logger: tail.DiscardingLogger})

	if err != nil {
		return false, err
	}

	for l := range t.Lines {
		if l.Text == desiredLine {
			return true, nil
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
			c.spinner.UpdateMessage("installing nix packages")
			continue
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			c.spinner.UpdateMessage("loading please wait")
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
