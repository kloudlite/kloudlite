package boxpkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
	"github.com/nxadm/tail"
)

var containerNotStartedErr = fmt.Errorf("container not started")

func (c *client) Start() error {
	c.spinner.Start("initiating container please wait")

	if c.verbose {
		fn.Logf("starting container in: %s", text.Blue(c.cwd))
	}

	cr, err := c.getContainer()
	if err != nil {
		return err
	}

	if cr.Name != "" {
		c.spinner.Stop()
		fn.Logf("container %s already running in %s", text.Yellow(cr.Name), text.Blue(cr.Path))

		if c.cwd != cr.Path {
			fn.Printf("do you want to stop that and start here? [Y/n]")
		} else {
			fn.Printf("do you want to restart it? [y/N]")
		}

		var response string
		_, _ = fmt.Scanln(&response)
		if c.cwd != cr.Path && response == "n" {
			return containerNotStartedErr
		}

		if c.cwd == cr.Path && response != "y" {
			return containerNotStartedErr
		}

		if err := c.Stop(); err != nil {
			return err
		}

		return c.Start()
	}

	if err := c.ensurePublicKey(); err != nil {
		return err
	}

	if err := c.ensureCacheExist(); err != nil {
		return err
	}

	envs, mmap, err := server.GetLoadMaps()
	if err != nil {
		return err
	}

	// local setup
	kConf, err := c.loadConfig(mmap, envs)
	if err != nil {
		return err
	}

	td, err := os.MkdirTemp("", "kl-tmp")
	if err != nil {
		return err
	}

	defer func() {
		os.RemoveAll(td)
	}()

	stdErrPath := fmt.Sprintf("%s/stderr.log", td)
	stdOutPath := fmt.Sprintf("%s/stdout.log", td)

	if err := c.startContainer(*kConf, td); err != nil {
		return err
	}

	if err != nil {
		return err
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

			fn.Log(text.Blue("container started successfully"))
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
			c.spinner.Update("installing nix packages")
			continue
		}

		if l.Text == "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE" {
			c.spinner.Update("loading please wait")
			continue
		}

		if c.verbose {
			switch stream {
			case "stderr":
				fn.Logf("%s: %s", text.Red("[stderr]"), l.Text)
			default:
				fn.Logf("%s: %s", text.Blue("[stdout]"), l.Text)
			}
		}
	}

	return false, nil
}
