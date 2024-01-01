package runner

import (
	"fmt"
	"os"
	"path"

	common_cmd "github.com/kloudlite/kl/cmd/common"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib"
	common_util "github.com/kloudlite/kl/lib/common"
	"github.com/kloudlite/kl/lib/server"
	"github.com/kloudlite/kl/lib/util"
	"github.com/spf13/cobra"
)

var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize your " + constants.CmdName + "-config file with some sample values",
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
				common_util.PrintError(e)
				return
			}

			initFile = &server.KLFileType{
				Version: "v1",
				Name:    path.Base(dir),
				Mres:    make([]server.ResType, 0),
				Configs: make([]server.ResType, 0),
				Secrets: make([]server.ResType, 0),
				Env:     []server.EnvType{{Key: "SAMPLE_ENV", Value: "sample_value"}},
				// Ports:   []string{},
				FileMount: server.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]server.FileEntry, 0),
				},
			}

		} else {
			fmt.Println("file already present")
		}

		accountId, _ := util.CurrentAccountName()

		if aId == "" && accountId == "" {
			acc, e := common_cmd.SelectAccount([]string{})

			if e != nil {
				common_util.PrintError(e)
				return
			}

			e = common_cmd.SelectAccount(acc)
			if e != nil {
				common_util.PrintError(e)
				return
			}

		}

		if aId != "" {
			e := lib.SelectAccount(aId)

			if e != nil {
				common_util.PrintError(e)
				return
			}

		}

		projectId, _ := server.CurrentProjectId()

		if pId == "" && projectId == "" {
			projectId, e := use.SelectProject([]string{})
			if e != nil {
				common_util.PrintError(e)
				return
			}

			e = lib.SelectProject(projectId)
			if e != nil {
				common_util.PrintError(e)
				return
			}
		}

		if pId != "" {
			// TODO
			e := lib.SelectProject(pId)

			if e != nil {
				common_util.PrintError(e)
				return
			}

		}

		err = server.WriteKLFile(*initFile)

		if err != nil {
			common_util.PrintError(err)
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
