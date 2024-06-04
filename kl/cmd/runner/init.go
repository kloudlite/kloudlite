package runner

import (
	"github.com/kloudlite/kl/domain/client"
	"github.com/kloudlite/kl/domain/server"
	fn "github.com/kloudlite/kl/pkg/functions"
	"github.com/kloudlite/kl/pkg/ui/fzf"

	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "initialize a kl-config file",
	Long:  `use this command to initialize a kl-config file`,

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

			packages := []string{"vim", "git"}

			defEnv := ""
			if len(envs) != 0 {
				de, err := fzf.FindOne(envs, func(item server.Env) string {
					return item.Metadata.Name
				}, fzf.WithPrompt("Select default environment >"))

				if err != nil {
					fn.PrintError(err)
					return
				}

				defEnv = de.Metadata.Name
			} else {
				fn.Warn("no environment found, please create environments from dashboard")
			}

			initFile = &client.KLFileType{
				Version:    "v1",
				DefaultEnv: defEnv,
				Packages:   packages,
				// Mres:       make([]client.ResType, 0),
				// Configs:    make([]client.ResType, 0),
				// Secrets:    make([]client.ResType, 0),
				EnvVars: []client.EnvType{{Key: "SAMPLE", Value: fn.Ptr("sampleValue")}},
				Mounts:  client.Mounts{},
			}
			if defEnv == "" {
				fn.Warn("No environment found, Please create environments from dashboard\n")
			} else {
				fn.Log("default env set to: ", defEnv)
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
	InitCommand.Flags().StringP("account", "a", "", "account name")
	InitCommand.Flags().StringP("file", "f", "", "file name")
	fn.WithKlFile(InitCommand)
}
