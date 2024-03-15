package domain_util

import (
	"os"
	"path"

	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
)

func MountEnv(args []string) error {
	klfile, err := client.GetKlFile("")
	if err != nil {
		return err
	}

	envs, mmap, err := server.GetLoadMaps()
	if err != nil {
		return err
	}

	mountfiles := map[string]string{}

	for _, fe := range klfile.FileMount.Mounts {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		mountfiles[pth] = mmap[pth]
	}

	if err = mounter.Mount(mountfiles, klfile.FileMount.MountBasePath); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	for _, ev := range klfile.Env {
		envs[ev.Key] = ev.Value
	}

	envs["KL_MOUNT_PATH"] = path.Join(cwd, klfile.FileMount.MountBasePath)

	if err = mounter.Load(envs, args[1:]); err != nil {
		return err
	}

	return nil
}
