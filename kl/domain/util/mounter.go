package domain_util

import (
	"os"
	"path"

	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/pkg/functions"
)

const (
	MountPath = "./.kl-mounts"
)

func MountEnv(args []string) error {
	klfile, err := client.GetKlFile("")
	if err != nil {
		return functions.NewE(err)
	}

	envs, mmap, err := server.GetLoadMaps()
	if err != nil {
		return functions.NewE(err)
	}

	mountfiles := map[string]string{}
	for _, fe := range klfile.Mounts.GetMounts() {
		pth := fe.Path
		if pth == "" {
			pth = fe.Key
		}

		mountfiles[pth] = mmap[pth]
	}

	if err = mounter.Mount(mountfiles, MountPath); err != nil {
		return functions.NewE(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	for _, ev := range klfile.EnvVars.GetEnvs() {
		envs[ev.Key] = ev.Value
	}

	envs["KL_MOUNT_PATH"] = path.Join(cwd, MountPath)

	if err = mounter.Load(envs, args[1:]); err != nil {
		return functions.NewE(err)
	}

	return nil
}
