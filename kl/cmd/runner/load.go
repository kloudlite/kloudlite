package runner

import (
	"os"
	"path"

	"github.com/kloudlite/kl/cmd/runner/mounter"
	"github.com/kloudlite/kl/constants"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/spf13/cobra"
)

var LoadCommand = &cobra.Command{
	Use:   "load",
	Short: "load environment variables and mount config files according to defined in " + constants.CmdName + "-config file",
	Long: `Load Environment
This command help you to load environments of the server according to you defined in your kl-config file.

Examples:
  # load environments and mount the configs
  kl load

	# load environments and execute a program with that loaded environments
	kl load <your_cmd>

	# example with npm start
	kl load npm start

	# get environments in json format
	kl load -o json

	# get environments in yaml format
	kl load -o yaml

	# start a new shell with loaded environments
	kl load shell

	# example of env with zsh shell
	kl load zsh
	`,
	Run: func(_ *cobra.Command, args []string) {
		loadEnv(args)
	},
}

func loadEnv(args []string) {
	generatedConfig, err := server.GenerateEnv()
	if err != nil {
		common_util.PrintError(err)
		return
	}

	klfile, err := server.GetKlFile(nil)
	if err != nil {
		common_util.PrintError(err)
		return
	}

	err = mounter.Mount(generatedConfig.MountFiles, klfile.FileMount.MountBasePath)

	if err != nil {
		common_util.PrintError(err)
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	generatedConfig.EnvVars["KL_MOUNT_PATH"] = path.Join(cwd, klfile.FileMount.MountBasePath)

	err = mounter.Load(generatedConfig.EnvVars, args)

	if err != nil {
		common_util.PrintError(err)
		return
	}

}
