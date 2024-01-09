package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kl",
	Short: "kl is command line interface to interact with kloudlite environments",
	Long: `
you can use kl -- <cmd> to execute any command with loaded environments.
# example
kl -- npm start

kl(Kloudlite Cli) will manage and attach to kloudlite environments.

Find more information at https://kloudlite.io/docs/cli

> NOTE: default kl-config file is kl.yml you can provide your own by providing KLCONFIG_PATH to the environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			return
		}

		klfile, err := client.GetKlFile(nil)
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

		if err = mounter.Load(envs, args); err != nil {
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
}
