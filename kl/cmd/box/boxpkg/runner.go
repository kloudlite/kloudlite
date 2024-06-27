package boxpkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/kloudlite/kl/pkg/functions"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/spinner"
	"github.com/kloudlite/kl/pkg/ui/text"
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

var NotFoundErr = functions.Error("container not found")

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
		return nil, functions.NewE(err)
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
		return nil, functions.NewE(err)
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

// func (c *client) waitForContReady(containerId string) error {
// 	timeoutCtx, cf := context.WithTimeout(context.TODO(), 1*time.Minute)
//
// 	cancelFn := func() {
// 		defer cf()
// 	}
//
// 	defer cancelFn()
//
// 	status := make(chan int, 1)
// 	go func() {
// 		ok, err := c.readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:CRASHED", "stderr", true)
// 		if err != nil {
// 			fn.PrintError(err)
// 			status <- 2
// 			cf()
// 			return
// 		}
// 		if ok {
// 			status <- 1
// 		}
// 	}()
//
// 	go func() {
// 		ok, err := c.readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:SETUP_COMPLETE", "stdout", true)
// 		if err != nil {
// 			fn.PrintError(err)
// 			status <- 2
// 			return
// 		}
//
// 		if ok {
// 			status <- 0
// 		}
// 	}()
//
// 	select {
// 	case exitCode := <-status:
// 		{
// 			spinner.Client.Stop()
// 			cancelFn()
// 			if exitCode != 0 {
// 				c.readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:SETUP_COMPLETE", "stdout", false)
// 				c.readTillLine(timeoutCtx, containerId, "kloudlite-entrypoint:CRASHED", "stderr", false)
// 				return fn.Error("failed to start container")
// 			}
//
// 			// functions.Log(text.Blue("container started successfully"))
// 		}
// 	}
//
// 	return nil
// }

func (c *client) readTillLine(ctx context.Context, containerId string, desiredLine, stream string, follow bool) (bool, error) {
	cout, err := c.cli.ContainerLogs(ctx, containerId, container.LogsOptions{
		ShowStdout: func() bool {
			return stream == "stdout"
		}(),
		ShowStderr: func() bool {
			return stream == "stderr"
		}(),
		Follow: follow,
		Since:  time.Now().Format(time.RFC3339),
	})
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(cout)

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) > 8 {
			line = line[8:]
		}

		if line == desiredLine {
			return true, nil
		}

		if line == "kloudlite-entrypoint:INSTALLING_PACKAGES" {
			spinner.Client.UpdateMessage("installing nix packages")
			continue
		}

		if line == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			spinner.Client.UpdateMessage("loading please wait")
			continue
		}

		if c.verbose {
			switch stream {
			case "stderr":
				fn.Logf("%s: %s", text.Yellow("[stderr]"), line)
			default:
				fn.Logf("%s: %s", text.Blue("[stdout]"), line)
			}
		}
	}

	return false, nil
}

func writeOnUserScope(fpath string, data []byte) error {
	if err := os.WriteFile(fpath, data, 0o644); err != nil {
		return functions.NewE(err)
	}

	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return functions.NewE(err)
		}
	}

	return nil
}

func userOwn(fpath string) error {
	if usr, ok := os.LookupEnv("SUDO_USER"); ok {
		if err := fn.ExecCmd(
			fmt.Sprintf("chown -R %s %s", usr, filepath.Dir(fpath)), nil, false,
		); err != nil {
			return functions.NewE(err)
		}
	}

	return nil
}
