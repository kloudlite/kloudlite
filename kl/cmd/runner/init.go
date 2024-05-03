package runner

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"

	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize your kl-config file with some sample values",
	Long: `kl init

This command initialize a kl-config.

Examples:
  # Initialize the kl-config file
  kl init
	`,

	Run: func(cmd *cobra.Command, _ []string) {

		aName := fn.ParseStringFlag(cmd, "account")
		filePath := fn.ParseKlFile(cmd)
		initFile, err := client.GetKlFile(filePath)

		if err != nil {

			envs, err := server.ListEnvs(fn.MakeOption("accountName", aName))
			if err != nil {
				fn.PrintError(err)
				return
			}

			packages := []string{"vim", "git", "go"}

			defEnv := ""
			if len(envs) != 0 {
				defEnv = envs[0].Metadata.Name
			}
			initFile = &client.KLFileType{
				Version:    "v1",
				DefaultEnv: defEnv,
				Packages:   packages,
				Mres:       make([]client.ResType, 0),
				Configs:    make([]client.ResType, 0),
				Secrets:    make([]client.ResType, 0),
				Env:        []client.EnvType{{Key: "SAMPLE", Value: "sampleValue"}},
				EnvVars:    make([]client.EnvType, 0),
				FileMount: client.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]client.FileEntry, 0),
				},
			}
			if defEnv == "" {
				fn.Warn("No environment found for the project, Please create environments from dashboard\n")
			} else {
				fn.Log("Default env set to: ", defEnv)
			}

		} else {
			fn.Log("file already present \n")
		}

		if err = client.WriteKLFile(*initFile); err != nil {
			fn.PrintError(err)
			return
		}

		fn.Log("Initialized file ", client.GetConfigPath())
	},
}

func init() {
	InitCommand.Flags().StringP("project", "p", "", "project name")
	InitCommand.Flags().StringP("account", "a", "", "account name")
	InitCommand.Flags().StringP("file", "f", "", "file name")
	fn.WithKlFile(InitCommand)
}
