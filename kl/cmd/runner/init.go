package runner

import (
	"fmt"
	util2 "github.com/kloudlite/kl/domain/client"
	fn "github.com/kloudlite/kl/pkg/functions"
	"os"
	"path"

	common_cmd "github.com/kloudlite/kl/cmd/common"
	"github.com/kloudlite/kl/cmd/use"
	"github.com/kloudlite/kl/constants"
	"github.com/kloudlite/kl/lib"
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

		pName := cmd.Flag("project").Value.String()
		aName := cmd.Flag("account").Value.String()

		initFile, err := util2.GetKlFile(nil)

		if err != nil {

			dir, e := os.Getwd()
			if e != nil {
				fn.PrintError(e)
				return
			}

			initFile = &util2.KLFileType{
				Version: "v1",
				Name:    path.Base(dir),
				Mres:    make([]util2.ResType, 0),
				Configs: make([]util2.ResType, 0),
				Secrets: make([]util2.ResType, 0),
				Env:     []util2.EnvType{{Key: "SAMPLE_ENV", Value: "sample_value"}},
				// Ports:   []string{},
				FileMount: util2.MountType{
					MountBasePath: "./.mounts",
					Mounts:        make([]util2.FileEntry, 0),
				},
			}

		} else {
			fmt.Println("file already present")
		}

		accountId, _ := util2.CurrentAccountName()

		if aName == "" && accountId == "" {
			_, e := common_cmd.SelectAccount([]string{})

			if e != nil {
				fn.PrintError(e)
				return
			}

		}

		if aName != "" {
			e := lib.SelectAccount(aName)

			if e != nil {
				fn.PrintError(e)
				return
			}

		}

		projectId, _ := util2.CurrentProjectName()

		if pName == "" && projectId == "" {
			projectId, e := use.SelectProject([]string{})
			if e != nil {
				fn.PrintError(e)
				return
			}

			e = lib.SelectProject(projectId)
			if e != nil {
				fn.PrintError(e)
				return
			}
		}

		if pName != "" {
			// TODO
			e := lib.SelectProject(pName)

			if e != nil {
				fn.PrintError(e)
				return
			}

		}

		err = util2.WriteKLFile(*initFile)

		if err != nil {
			fn.PrintError(err)
			return
		}

		fmt.Println("Initialized file", util2.GetConfigPath())
	},
}

func init() {
	p := ""
	a := ""

	InitCommand.Flags().StringVarP(&p, "projectId", "p", "", "project id")
	InitCommand.Flags().StringVarP(&a, "accountId", "a", "", "account id")
}
