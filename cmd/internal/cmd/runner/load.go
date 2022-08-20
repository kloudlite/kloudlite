package runner

import (
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/cmd/runner/mounter"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var LoadCommand = &cobra.Command{
	Use:   "load",
	Short: "load environment variables and mount config files according to defined in " + constants.CMD_NAME + "-config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		loadEnv(args)
	},
}

func loadEnv(args []string) {
	generatedConfig, err := server.GenerateEnv()
	if err != nil {
		common.PrintError(err)
		return
	}

	klfile, err := server.GetKlFile(nil)
	if err != nil {
		common.PrintError(err)
		return
	}

	err = mounter.Mount(generatedConfig.MountFiles, klfile.FileMount.MountBasePath)

	if err != nil {
		common.PrintError(err)
		return
	}

	generatedConfig.EnvVars["KL_MOUNT_PATH"] = klfile.FileMount.MountBasePath

	err = mounter.Load(generatedConfig.EnvVars, args)

	if err != nil {
		common.PrintError(err)
		return
	}

}
