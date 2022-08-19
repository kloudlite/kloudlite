package add

import (
	"fmt"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"kloudlite.io/cmd/internal/common"
	"kloudlite.io/cmd/internal/constants"
	"kloudlite.io/cmd/internal/lib/server"
)

var addConfigCommand = &cobra.Command{
	Use:   "config",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		SelectConfig()
	},
}

func SelectConfig() {

	klFile, err := server.GetKlFile(nil)
	if err != nil {
		common.PrintError(err)
		es := "Please run '" + constants.CMD_NAME + " init' if you are not initialized the file already"
		common.PrintError(fmt.Errorf(es))
		return
	}

	configs, err := server.GetConfigs()

	if err != nil {
		common.PrintError(err)
		return
	}

	if len(configs) == 0 {
		es := "No configs created yet on server"
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedGroupIndex, err := fuzzyfinder.Find(
		configs,
		func(i int) string {
			return configs[i].Name
		},
		fuzzyfinder.WithPromptString("Select Config Group >"),
	)

	if err != nil {
		common.PrintError(err)
	}

	selectedConfigGroup := configs[selectedGroupIndex]

	if len(selectedConfigGroup.Entries) == 0 {
		es := fmt.Sprintf("No configs added yet to %s config", selectedConfigGroup.Name)
		common.PrintError(fmt.Errorf(es))
		return
	}

	selectedKeyIndex, err := fuzzyfinder.Find(
		selectedConfigGroup.Entries,
		func(i int) string {
			return selectedConfigGroup.Entries[i].Key
		},
		fuzzyfinder.WithPromptString(fmt.Sprintf("Select Key of %s >", selectedConfigGroup.Name)),
	)
	if err != nil {
		common.PrintError(err)
	}

	selectedConfigKey := selectedConfigGroup.Entries[selectedKeyIndex]

	matchedGroupIndex := -1
	for i, rt := range klFile.Configs {
		if rt.Id == selectedConfigGroup.Id {
			matchedGroupIndex = i
			break
		}
	}

	if matchedGroupIndex != -1 {
		matchedKeyIndex := -1

		for i, ret := range klFile.Configs[matchedGroupIndex].Env {
			if ret.RefKey == selectedConfigKey.Key {
				matchedKeyIndex = i
				break
			}
		}

		if matchedKeyIndex == -1 {
			klFile.Configs[matchedGroupIndex].Env = append(klFile.Configs[matchedGroupIndex].Env, server.ResEnvType{
				Key:    selectedConfigKey.Key,
				RefKey: selectedConfigKey.Key,
			})
		}
	} else {
		klFile.Configs = append(klFile.Configs, server.ResType{
			Id:   selectedConfigGroup.Id,
			Name: selectedConfigGroup.Name,
			Env: []server.ResEnvType{
				{
					Key:    selectedConfigKey.Key,
					RefKey: selectedConfigKey.Key,
				},
			},
		})
	}

	err = server.WriteKLFile(*klFile)
	if err != nil {
		common.PrintError(err)
	}

	fmt.Printf("added config %s/%s to your %s-file\n", selectedConfigGroup.Name, selectedConfigKey.Key, constants.CMD_NAME)
}
