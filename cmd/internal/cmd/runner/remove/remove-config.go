package remove

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var removeConfigCommand = &cobra.Command{
	Use:   "config",
	Short: "remove one config environment from your " + constants.CMD_NAME + "-config",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command.`,
	Run: func(cmd *cobra.Command, args []string) {
		removeConfig()
	},
}

func removeConfig() {

	klFile, err := server.GetKlFile(nil)
	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	if len(klFile.Configs) == 0 {
		es := "No configs added yet in your file"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedConfigIndex, err := fuzzyfinder.Find(
		klFile.Configs,
		func(i int) string {
			return fmt.Sprintf("%s, %s", klFile.Configs[i].Name, klFile.Configs[i].Id)
		},
		fuzzyfinder.WithPromptString("Select Config Group >"),
	)

	if err != nil {
		common.PrintError(err)
		return
	}

	selectedConfig := klFile.Configs[selectedConfigIndex]

	if len(selectedConfig.Env) == 1 {
		newConfigs := make([]server.ResType, 0)
		for i, rt := range klFile.Configs {
			if i == selectedConfigIndex {
				continue
			}
			newConfigs = append(newConfigs, rt)
		}

		klFile.Configs = newConfigs

		fmt.Printf("removed config %s form your %s-file\n", selectedConfig.Name, constants.CMD_NAME)

	} else {

		selectedKeyIndex, e := fuzzyfinder.Find(
			selectedConfig.Env,
			func(i int) string {
				return fmt.Sprintf("%s, %s", selectedConfig.Env[i].Key, selectedConfig.Env[i].RefKey)
			},
			fuzzyfinder.WithPromptString("Select Config Key >"),
		)

		if e != nil {
			common.PrintError(e)
			return
		}

		newEnvs := make([]server.ResEnvType, 0)
		for i, ret := range selectedConfig.Env {
			if i == selectedKeyIndex {
				continue
			}
			newEnvs = append(newEnvs, ret)
		}

		klFile.Configs[selectedConfigIndex].Env = newEnvs

		fmt.Printf("removed key %s/%s form your %s-file\n", selectedConfig.Name, selectedConfig.Name, constants.CMD_NAME)
	}

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common.PrintError(err)
	}

}
