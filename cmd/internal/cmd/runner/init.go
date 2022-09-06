package runner

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/cmd/use"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib"
	"kloudlite.io/cmd/internal/lib/server"
)

// initCmd represents the init command
var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize your " + constants.CMD_NAME + "-config file with some sample values",
	Long: `kl init

This command initialize a kl-config.

Examples:
  # Initialize the kl-config file
  kl init
	`,

	Run: func(cmd *cobra.Command, _ []string) {

		pId := cmd.Flag("projectId").Value.String()
		aId := cmd.Flag("accountId").Value.String()

		initFile, err := server.GetKlFile(nil)

		if err != nil {

			dir, e := os.Getwd()
			if e != nil {
				common.PrintError(e)
				return
			}

			initFile = &server.KLFileType{
				Version: "v1",
				Name:    path.Base(dir),
				Mres:    make([]server.ResType, 0),
				Configs: make([]server.ResType, 0),
				Secrets: make([]server.ResType, 0),
				Env:     []server.EnvType{{Key: "SAMPLE_ENV", Value: "sample_value"}},
				Ports:   []string{},
				FileMount: server.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]server.FileEntry, 0),
				},
			}

		} else {
			fmt.Println("file already present")
		}

		accountId, _ := server.CurrentAccountId()

		if aId == "" && accountId == "" {
			accountId, e := use.SelectAccount([]string{})

			if e != nil {
				common.PrintError(e)
				return
			}

			e = lib.SelectAccount(accountId)
			if e != nil {
				common.PrintError(e)
				return
			}

		}

		if aId != "" {
			e := lib.SelectAccount(aId)

			if e != nil {
				common.PrintError(e)
				return
			}

		}

		projectId, _ := server.CurrentProjectId()

		if pId == "" && projectId == "" {
			projectId, e := use.SelectProject([]string{})
			if e != nil {
				common.PrintError(e)
				return
			}

			e = lib.SelectProject(projectId)
			if e != nil {
				common.PrintError(e)
				return
			}
		}

		if pId != "" {
			// TODO
			e := lib.SelectProject(pId)

			if e != nil {
				common.PrintError(e)
				return
			}

		}

		err = server.WriteKLFile(*initFile)

		if err != nil {
			common.PrintError(err)
			return
		}

		fmt.Println("Initialized file", server.GetConfigPath())
	},
}

func init() {
	p := ""
	a := ""

	InitCommand.Flags().StringVarP(&p, "projectId", "p", "", "project id")
	InitCommand.Flags().StringVarP(&a, "accountId", "a", "", "account id")
}
