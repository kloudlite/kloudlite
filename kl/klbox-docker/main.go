package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/klbox-docker/devboxfile"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/text"
)

func main() {
	if err := Run(); err != nil {
		fn.PrintError(err)
		os.Exit(1)
	}
}

func Run() error {
	var configFile string
	flag.StringVar(&configFile, "conf", "", "--conf /path/to/config.json")
	flag.Parse()

	if configFile == "" {
		return fmt.Errorf("no config file provided")
	}

	b, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var c devboxfile.DevboxConfig
	err = json.Unmarshal(b, &c)
	if err != nil {
		return err
	}

	for k, v := range c.KlConfig.Mounts {
		if err := os.MkdirAll(filepath.Dir(k), fs.ModePerm); err != nil {
			return err
		}

		if err := os.Chown(filepath.Dir(k), 1000, 1000); err != nil {
			return err
		}

		if err := os.WriteFile(k, []byte(v), fs.ModePerm); err != nil {
			return err
		}

		if err := os.Chown(k, 1000, 1000); err != nil {
			return err
		}
	}

	wgPath := "/etc/resolv.conf"
	if c.KlConfig.Dns != "" {
		if err := os.WriteFile(wgPath, []byte(c.KlConfig.Dns), fs.ModePerm); err != nil {
			return err
		}
	}

	kt, err := client.GetKlFile(path.Join("/home/kl/workspace", "kl.yml"))
	if err != nil {
		fn.PrintError(err)
		return nil
	}

	if len(kt.InitScripts) > 0 {
		fn.Log(text.Blue("[#] Running init scripts"))
		defer fn.Log(text.Blue("[#] Finished running init scripts"))
	}

	for _, v := range kt.InitScripts {
		if err := RunScript(v); err != nil {
			fn.PrintError(fmt.Errorf("error running init script: %q", v))
		}
	}

	return nil
}

func RunScript(script string) error {
	username := "kl"

	// Lookup the user
	usr, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("Error looking up user: %s", err.Error())
	}

	// Get the UID and GID
	uid := usr.Uid
	gid := usr.Gid

	// Convert UID and GID to integers
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return fmt.Errorf("Invalid UID: %s", err.Error())
	}
	gidInt, err := strconv.Atoi(gid)
	if err != nil {
		return fmt.Errorf("Invalid GID: %s", err)
	}

	// Prepare the command
	cmd := exec.Command("bash", "-c", script)

	// Set the UID and GID for the command
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uidInt),
			Gid: uint32(gidInt),
		},
	}

	fn.Logf("[%s] %s", text.Blue("+"), script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "/home/kl/workspace"

	return cmd.Run()
}
