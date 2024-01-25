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

func GetRootHelp(cmd *cobra.Command) string {

	return fmt.Sprintf(`Usage: kl [command] [options] [-- command] [command options]

These are common kl commands used in various situatiions:

Authentication
      auth login                  - login to kloudlite account
      auth logout                 - logout from kloudlite account
      auth whoami                 - get current logged in user

Manage Context:
      auth login                  - login to kloudlite account
      auth logout                 - logout from kloudlite account
      status                      - get status of your current context (user, account, project, environment, vpn status)

      list account                - list all accounts
      switch account              - switch between kloudlite accounts

Setup a kloudlite environment:
      init                        - initilize kloudlite configuration file in current working directory
      add config                  - add config entries to kloudlite configuration file
      add secret                  - add secret entries to kloudlite cofiguration file
      add mres                    - add managed resource params to kloudlite configuration file

Working inside environment:
      intercept <appname>         - intercept the application in the environment with your device.
                                    This will tunnel all the incoming traffic to your device

      -- <command>                - execute any command with loaded env variables
                                    Example: kl -- npm start

      list env                    - list all environments in current project
      switch environment          - inside the current project context you can switch between environments

VPN management:
      vpn connect                 - connect/switch your device to current working environment (requires sudo)
      vpn disconnect              - disconnect your device (requires sudo)
      vpn wg                      - get status of your device. (handshake, ip, port, etc) (requires sudo)
      vpn status                  - get status of your device. (connected/disconnected & connected environemnt)
      vpn expose                  - expose your local device ports

Fetch resources of current environment:
      get config <config-name>    - get config entries
      get secret <secret-name>    - get secret entries
      get mres <mres-name>        - get managed resource parameters
      list config                 - list all configs in current environment of project
      list secret                 - list all secrets in current environemnt of project
      list mres                   - list all managed resources in current environemnt of project
	`)
}

var Version = "development"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:                "kl",
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 || args[0] != "--" {

			s := GetRootHelp(cmd)
			fn.Log(s)

			// if err := cmd.Help(); err != nil {
			// 	fn.Log(err)
			// 	os.Exit(1)
			// }
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
	rootCmd.Version = Version
}
