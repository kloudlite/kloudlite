package runner

import (
	"fmt"

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

		pName := cmd.Flag("project").Value.String()
		aName := cmd.Flag("account").Value.String()

		initFile, err := client.GetKlFile(nil)

		if err != nil {

			acc, err := server.EnsureAccount(aName)
			if err != nil {
				fn.PrintError(err)
				return
			}

			prj, err := server.EnsureProject(
				[]fn.Option{
					fn.MakeOption("accountName", aName),
					fn.MakeOption("projectName", pName),
				}...,
			)
			if err != nil {
				fn.PrintError(err)
				return
			}

			initFile = &client.KLFileType{
				Version: "v1",
				Project: fmt.Sprintf("%s/%s", acc, prj),
				Mres:    make([]client.ResType, 0),
				Configs: make([]client.ResType, 0),
				Secrets: make([]client.ResType, 0),
				Env:     []client.EnvType{{Key: "SAMPLE_ENV", Value: "sample_value"}},
				FileMount: client.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]client.FileEntry, 0),
				},
			}
		} else {
			fmt.Println("file already present")
		}

		if err = client.WriteKLFile(*initFile); err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println("Initialized file", client.GetConfigPath())
	},
}

func init() {
	p := ""
	a := ""

	InitCommand.Flags().StringVarP(&p, "project", "p", "", "project name")
	InitCommand.Flags().StringVarP(&a, "account", "a", "", "account name")
}
