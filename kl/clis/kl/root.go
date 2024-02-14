package kl

import (
	"os"
	"path"

	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	"github.com/kloudlite/kl/flags"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                flags.CliName,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		if (len(args) != 0) && (args[0] == "--version" || args[0] == "-v") {
			fn.Log(cmd.Version)
			return
		}

		if len(args) < 2 || args[0] != "--" {
			// fn.Log(GetRootHelp(cmd))

			if err := cmd.Help(); err != nil {
				fn.Log(err)
				os.Exit(1)
			}
			return
		}

		klfile, err := client.GetKlFile("")
		if err != nil {
			fn.PrintError(err)
			return
		}

		envs, cmap, smap, err := server.GetLoadMaps()
		if err != nil {
			fn.PrintError(err)
			return
		}

		mountfiles := map[string]string{}

		for _, fe := range klfile.FileMount.Mounts {
			pth := fe.Path
			if pth == "" {
				pth = fe.Key
			}

			if fe.Type == client.ConfigType {
				mountfiles[pth] = cmap[fe.Name][fe.Key].Value
			} else {
				mountfiles[pth] = smap[fe.Name][fe.Key].Value
			}
		}

		if err = mounter.Mount(mountfiles, klfile.FileMount.MountBasePath); err != nil {
			fn.PrintError(err)
			return
		}

		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}

		envs["KL_MOUNT_PATH"] = path.Join(cwd, klfile.FileMount.MountBasePath)

		if err = mounter.Load(envs, args[1:]); err != nil {
			fn.PrintError(err)
			return
		}

	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = flags.Version
}
